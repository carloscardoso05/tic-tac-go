package event

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type EventType string

const (
	EventCreate       EventType = "create"
	EventJoin         EventType = "join"
	EventMark         EventType = "mark"
	EventLeave        EventType = "leave"
	EventRoomCreated  EventType = "room_created"
	EventUserJoined   EventType = "user_joined"
	EventUserLeft     EventType = "user_left"
	EventTileMarked   EventType = "tile_marked"
	EventGameEnded    EventType = "game_ended"
	EventErrorOccurred EventType = "error_occurred"
)

type Client struct {
	ID   uuid.UUID
	Name string
	Conn *websocket.Conn
}

func NewClient(conn *websocket.Conn, name string) *Client {
	return &Client{
		ID:   uuid.New(),
		Name: name,
		Conn: conn,
	}
}

type Message struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

type RoomCreated struct {
	Room string `json:"room" validate:"required"`
}

type TileMarked struct {
	Room string `json:"room" validate:"required"`
	User string `json:"user" validate:"required"`
	Tile int    `json:"tile" validate:"required"`
}

type UserJoined struct {
	Room string `json:"room" validate:"required"`
	User string `json:"user" validate:"required"`
}

type UserLeft struct {
	Room string `json:"room" validate:"required"`
	User string `json:"user" validate:"required"`
}

type GameEnded struct {
	Room   string `json:"room" validate:"required"`
	Winner string `json:"winner"`
}

type ErrorOccurred struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func NewMessage(eventType EventType, event any) Message {
	data, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	return Message{Type: eventType, Data: data}
}

var validate = validator.New()

func Parse[T any](data json.RawMessage) (*T, error) {
	event := new(T)
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	if err := validate.Struct(event); err != nil {
		return nil, err
	}
	return event, nil
}

func Read(ctx context.Context, conn *websocket.Conn) (Message, error) {
	var msg Message
	return msg, wsjson.Read(ctx, conn, &msg)
}

func Write(ctx context.Context, conn *websocket.Conn, msg Message) error {
	return wsjson.Write(ctx, conn, msg)
}

func WriteError(ctx context.Context, conn *websocket.Conn, err error) {
	sendErr := Write(ctx, conn, NewMessage(EventErrorOccurred, ErrorOccurred{Error: err.Error()}))
	if sendErr != nil {
		panic(sendErr)
	}
}

