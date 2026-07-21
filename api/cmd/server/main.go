package main

import (
	"api/domain/room"
	"api/event"
	"context"
	"fmt"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Player struct {
	ID   uuid.UUID
	Name string
	ctx  context.Context
	conn *websocket.Conn
}

type Hub struct {
	rooms map[uuid.UUID]room.Room
}

func main() {
	router := gin.Default()

	router.GET("/rooms", func(ctx *gin.Context) {
		conn, err := websocket.Accept(ctx.Writer, ctx.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "Bye")

		wrapper := new(event.Wrapper)
		for {
			/*
				1. ler e validar wrapper -> w := EventWrapper
				2. identificar tipo do evento
				3. fazer um switch case: para cada tipo, fazer o parse correto e realizar uma ação
			*/
			err := event.ParseWrapper(ctx, conn, wrapper)
			if err != nil {
				return
			}

			switch wrapper.Type {
			// case event.EventTypeErrorOcurred:
			case event.EventTypeGameEnded:
				event := event.ParseEvent[event.GameEnded](ctx, conn, wrapper)
				if event == nil {
					continue
				}

			case event.EventTypeTileMarked:
			case event.EventTypeUserJoined:
			case event.EventTypeUserLeft:
			default:
				wrapper := event.NewWrapper(
					event.EventTypeErrorOcurred,
					event.ErrorOcurred{Error: fmt.Errorf("EventType %s is invalid", wrapper.Type)},
				)
				event.Send(ctx, conn, wrapper)
			}
		}
	})

	router.Run()
}
