package transport

import (
	"api/event"
	"context"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func Accept(ctx *gin.Context) (*websocket.Conn, error) {
	return websocket.Accept(ctx.Writer, ctx.Request, nil)
}

func Read(ctx context.Context, conn *websocket.Conn) (event.Message, error) {
	return event.Read(ctx, conn)
}

func Write(ctx context.Context, conn *websocket.Conn, msg event.Message) error {
	return event.Write(ctx, conn, msg)
}

func WriteError(ctx context.Context, conn *websocket.Conn, err error) {
	event.WriteError(ctx, conn, err)
}
