package lobby

import (
	"api/internal/event"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

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

func (l *Lobby) HandleCreate(ctx context.Context, cl *event.Client, data json.RawMessage) error {
	type createPayload struct {
		Name string `json:"name" validate:"required"`
	}
	payload, err := event.Parse[createPayload](data)
	if err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	cl.Name = payload.Name

	m := &Match{ID: uuid.New(), P1: cl}

	l.mu.Lock()
	l.matches[m.ID] = m
	l.clients[cl.ID] = m
	l.mu.Unlock()

	msg := event.NewMessage(event.EventRoomCreated, event.RoomCreated{Room: m.ID.String()})
	return event.Write(ctx, cl.Conn, msg)
}

func (l *Lobby) HandleJoin(ctx context.Context, cl *event.Client, data json.RawMessage) error {
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
	m, ok := l.matches[roomID]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	l.mu.Lock()
	m.P2 = cl
	m.current = 0
	l.clients[cl.ID] = m
	l.mu.Unlock()

	joinedMsg := event.NewMessage(event.EventUserJoined, event.UserJoined{
		Room: m.ID.String(),
		User: cl.Name,
	})
	l.broadcast(ctx, m, joinedMsg)

	return nil
}

func (l *Lobby) HandleMark(ctx context.Context, cl *event.Client, data json.RawMessage) error {
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
	m, ok := l.matches[uuid.MustParse(payload.Room)]
	l.mu.Unlock()
	if !ok {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("room not found"))
		return nil
	}

	if m.P2 == nil {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("game not started yet"))
		return nil
	}

	if m.CurrentClient().ID != cl.ID {
		event.WriteError(ctx, cl.Conn, fmt.Errorf("not your turn"))
		return nil
	}

	if err := m.Mark(Tile(payload.Tile)); err != nil {
		event.WriteError(ctx, cl.Conn, err)
		return nil
	}

	tileMsg := event.NewMessage(event.EventTileMarked, event.TileMarked{
		Room: m.ID.String(),
		User: cl.Name,
		Tile: payload.Tile,
	})
	l.broadcast(ctx, m, tileMsg)

	status := m.Status()
	switch status {
	case P1_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   m.ID.String(),
			Winner: m.P1.Name,
		})
		l.broadcast(ctx, m, endMsg)
	case P2_WON:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   m.ID.String(),
			Winner: m.P2.Name,
		})
		l.broadcast(ctx, m, endMsg)
	case DRAW:
		endMsg := event.NewMessage(event.EventGameEnded, event.GameEnded{
			Room:   m.ID.String(),
			Winner: "",
		})
		l.broadcast(ctx, m, endMsg)
	}

	return nil
}

func (l *Lobby) HandleLeave(ctx context.Context, cl *event.Client, data json.RawMessage) error {
	l.removeClient(ctx, cl)
	return nil
}

func (l *Lobby) Disconnect(cl *event.Client) {
	l.removeClient(context.Background(), cl)
}

func (l *Lobby) removeClient(ctx context.Context, cl *event.Client) {
	l.mu.Lock()
	m := l.clients[cl.ID]
	delete(l.clients, cl.ID)
	if m != nil {
		delete(l.matches, m.ID)
	}
	l.mu.Unlock()

	if m != nil {
		leftMsg := event.NewMessage(event.EventUserLeft, event.UserLeft{
			Room: m.ID.String(),
			User: cl.Name,
		})
		l.broadcastExcept(ctx, m, cl, leftMsg)
	}
}

func (l *Lobby) broadcast(ctx context.Context, m *Match, msg event.Message) {
	if m.P1 != nil {
		event.Write(ctx, m.P1.Conn, msg)
	}
	if m.P2 != nil {
		event.Write(ctx, m.P2.Conn, msg)
	}
}

func (l *Lobby) broadcastExcept(ctx context.Context, m *Match, sender *event.Client, msg event.Message) {
	if m.P1 != nil && m.P1.ID != sender.ID {
		event.Write(ctx, m.P1.Conn, msg)
	}
	if m.P2 != nil && m.P2.ID != sender.ID {
		event.Write(ctx, m.P2.Conn, msg)
	}
}
