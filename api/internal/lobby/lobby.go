package lobby

import (
	"api/event"
	"api/internal/client"
	"api/internal/room"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Match struct {
	ID   uuid.UUID
	Game *room.Room
	P1   *client.Client
	P2   *client.Client
	mu   sync.Mutex
}

type Lobby struct {
	matches map[uuid.UUID]*Match
	clients map[uuid.UUID]*Match
	mu      sync.Mutex
}

func New() *Lobby {
	return &Lobby{
		matches: make(map[uuid.UUID]*Match),
		clients: make(map[uuid.UUID]*Match),
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

	match := &Match{
		ID: uuid.New(),
		P1: cl,
	}

	l.mu.Lock()
	l.matches[match.ID] = match
	l.clients[cl.ID] = match
	l.mu.Unlock()

	msg := event.NewMessage(event.EventRoomCreated, event.RoomCreated{Room: match.ID.String()})
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
	match, ok := l.matches[roomID]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	if match.P2 != nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room is full"))
		return nil
	}
	if match.P1 != nil && match.P1.ID == cl.ID {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("cannot join your own room"))
		return nil
	}

	match.P2 = cl
	l.mu.Lock()
	l.clients[cl.ID] = match
	l.mu.Unlock()

	p1 := room.Player{Name: match.P1.Name}
	p2 := room.Player{Name: cl.Name}
	match.Game = room.NewRoom(p1, p2)

	joinedMsg := event.NewMessage(event.EventUserJoined, event.UserJoined{
		Room: match.ID.String(),
		User: cl.Name,
	})
	l.broadcast(ctx, match, joinedMsg)

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
	match, ok := l.matches[uuid.MustParse(payload.Room)]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	match.mu.Lock()
	defer match.mu.Unlock()

	if match.Game == nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("game not started yet"))
		return nil
	}

	if match.Game.CurrentPlayer().Name != cl.Name {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("not your turn"))
		return nil
	}

	tile := room.Tile(payload.Tile)
	if err := match.Game.Mark(tile); err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	tileMsg := event.NewMessage(event.EventTileMarked, event.TileMarked{
		Room: match.ID.String(),
		User: cl.Name,
		Tile: payload.Tile,
	})
	l.broadcast(ctx, match, tileMsg)

	status := match.Game.Status()
	switch status {
	case room.P1_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   match.ID.String(),
			Winner: match.P1.Name,
		})
		l.broadcast(ctx, match, endMsg)
	case room.P2_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   match.ID.String(),
			Winner: match.P2.Name,
		})
		l.broadcast(ctx, match, endMsg)
	case room.DRAW:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   match.ID.String(),
			Winner: "",
		})
		l.broadcast(ctx, match, endMsg)
	}

	return nil
}

func (l *Lobby) HandleLeave(ctx context.Context, cl *client.Client, data json.RawMessage) error {
	l.removeClient(ctx, cl)
	return nil
}

func (l *Lobby) Disconnect(cl *client.Client) {
	l.removeClient(context.Background(), cl)
}

func (l *Lobby) removeClient(ctx context.Context, cl *client.Client) {
	l.mu.Lock()
	match := l.clients[cl.ID]
	delete(l.clients, cl.ID)
	if match != nil {
		delete(l.matches, match.ID)
	}
	l.mu.Unlock()

	if match != nil {
		leftMsg := event.NewMessage(event.EventUserLeft, event.UserLeft{
			Room: match.ID.String(),
			User: cl.Name,
		})
		l.broadcastExcept(ctx, match, cl, leftMsg)
	}
}

func (l *Lobby) broadcast(ctx context.Context, match *Match, msg event.Message) {
	if match.P1 != nil {
		event.Write(ctx, match.P1.Conn, msg)
	}
	if match.P2 != nil {
		event.Write(ctx, match.P2.Conn, msg)
	}
}

func (l *Lobby) broadcastExcept(ctx context.Context, match *Match, sender *client.Client, msg event.Message) {
	if match.P1 != nil && match.P1.ID != sender.ID {
		event.Write(ctx, match.P1.Conn, msg)
	}
	if match.P2 != nil && match.P2.ID != sender.ID {
		event.Write(ctx, match.P2.Conn, msg)
	}
}
