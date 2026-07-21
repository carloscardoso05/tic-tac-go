package lobby

import (
	"api/event"
	"api/internal/client"
	domainroom "api/internal/domain/room"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	ID   uuid.UUID
	Game *domainroom.Room
	P1   *client.Client
	P2   *client.Client
	mu   sync.Mutex
}

type Lobby struct {
	rooms   map[uuid.UUID]*Room
	clients map[uuid.UUID]*Room
	mu      sync.Mutex
}

func New() *Lobby {
	return &Lobby{
		rooms:   make(map[uuid.UUID]*Room),
		clients: make(map[uuid.UUID]*Room),
	}
}

func (l *Lobby) HandleCreate(ctx context.Context, cl *client.Client, data json.RawMessage) error {
	type createPayload struct {
		Name string `json:"name" validate:"required"`
	}
	payload, err := event.Parse[createPayload](data)
	if err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	cl.Name = payload.Name

	room := &Room{
		ID: uuid.New(),
		P1: cl,
	}

	l.mu.Lock()
	l.rooms[room.ID] = room
	l.clients[cl.ID] = room
	l.mu.Unlock()

	msg := event.NewMessage(event.EventRoomCreated, event.RoomCreated{Room: room.ID.String()})
	return event.Write(ctx, cl.Conn, msg)
}

func (l *Lobby) HandleJoin(ctx context.Context, cl *client.Client, data json.RawMessage) error {
	type joinPayload struct {
		Name string `json:"name" validate:"required"`
		Room string `json:"room" validate:"required"`
	}
	payload, err := event.Parse[joinPayload](data)
	if err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	cl.Name = payload.Name

	roomID, err := uuid.Parse(payload.Room)
	if err != nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("invalid room id"))
		return nil
	}

	l.mu.Lock()
	room, ok := l.rooms[roomID]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.P2 != nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room is full"))
		return nil
	}
	if room.P1 != nil && room.P1.ID == cl.ID {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("cannot join your own room"))
		return nil
	}

	room.P2 = cl
	l.mu.Lock()
	l.clients[cl.ID] = room
	l.mu.Unlock()

	p1 := domainroom.Player{Name: room.P1.Name}
	p2 := domainroom.Player{Name: cl.Name}
	room.Game = domainroom.NewRoom(p1, p2)

	joinedMsg := event.NewMessage(event.EventUserJoined, event.UserJoined{
		Room: room.ID.String(),
		User: cl.Name,
	})
	l.broadcast(ctx, room, joinedMsg)

	return nil
}

func (l *Lobby) HandleMark(ctx context.Context, cl *client.Client, data json.RawMessage) error {
	type markPayload struct {
		Room string `json:"room" validate:"required"`
		Tile int    `json:"tile" validate:"required"`
	}
	payload, err := event.Parse[markPayload](data)
	if err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	l.mu.Lock()
	room, ok := l.rooms[uuid.MustParse(payload.Room)]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if room.Game == nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("game not started yet"))
		return nil
	}

	if room.Game.CurrentPlayer().Name != cl.Name {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("not your turn"))
		return nil
	}

	tile := domainroom.Tile(payload.Tile)
	if err := room.Game.Mark(tile); err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	tileMsg := event.NewMessage(event.EventTileMarked, event.TileMarked{
		Room: room.ID.String(),
		User: cl.Name,
		Tile: payload.Tile,
	})
	l.broadcast(ctx, room, tileMsg)

	status := room.Game.Status()
	switch status {
	case domainroom.P1_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   room.ID.String(),
			Winner: room.P1.Name,
		})
		l.broadcast(ctx, room, endMsg)
	case domainroom.P2_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   room.ID.String(),
			Winner: room.P2.Name,
		})
		l.broadcast(ctx, room, endMsg)
	case domainroom.DRAW:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   room.ID.String(),
			Winner: "",
		})
		l.broadcast(ctx, room, endMsg)
	}

	return nil
}

func (l *Lobby) HandleLeave(ctx context.Context, cl *client.Client, data json.RawMessage) error {
	l.mu.Lock()
	room := l.clients[cl.ID]
	delete(l.clients, cl.ID)
	if room != nil {
		delete(l.rooms, room.ID)
	}
	l.mu.Unlock()

	if room != nil {
		leftMsg := event.NewMessage(event.EventUserLeft, event.UserLeft{
			Room: room.ID.String(),
			User: cl.Name,
		})
		l.broadcastExcept(ctx, room, cl, leftMsg)
	}

	return nil
}

func (l *Lobby) broadcast(ctx context.Context, room *Room, msg event.Message) {
	if room.P1 != nil {
		event.Write(ctx, room.P1.Conn, msg)
	}
	if room.P2 != nil {
		event.Write(ctx, room.P2.Conn, msg)
	}
}

func (l *Lobby) broadcastExcept(ctx context.Context, room *Room, sender *client.Client, msg event.Message) {
	if room.P1 != nil && room.P1.ID != sender.ID {
		event.Write(ctx, room.P1.Conn, msg)
	}
	if room.P2 != nil && room.P2.ID != sender.ID {
		event.Write(ctx, room.P2.Conn, msg)
	}
}
