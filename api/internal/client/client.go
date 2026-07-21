package client

import (
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Client struct {
	ID   uuid.UUID
	Name string
	Conn *websocket.Conn
}

func New(conn *websocket.Conn, name string) *Client {
	return &Client{
		ID:   uuid.New(),
		Name: name,
		Conn: conn,
	}
}
