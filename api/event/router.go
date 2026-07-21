package event

import (
	"api/internal/client"
	"context"
	"encoding/json"
	"fmt"
)

type Handler func(ctx context.Context, cl *client.Client, data json.RawMessage) error

type Router struct {
	handlers map[EventType]Handler
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[EventType]Handler),
	}
}

func (r *Router) Handle(eventType EventType, h Handler) {
	r.handlers[eventType] = h
}

func (r *Router) Route(ctx context.Context, cl *client.Client, msg Message) error {
	handler, ok := r.handlers[msg.Type]
	if !ok {
		WriteError(ctx, cl.Conn, fmt.Errorf("unknown event type: %s", msg.Type))
		return nil
	}
	return handler(ctx, cl, msg.Data)
}
