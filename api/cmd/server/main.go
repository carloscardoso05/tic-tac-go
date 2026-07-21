package main

import (
	"api/internal/event"
	"api/internal/lobby"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	l := lobby.New()
	router := setupRouter(l)

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		conn, err := websocket.Accept(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		handleConn(c, conn, router, l)
		conn.CloseNow()
	})

	r.Run()
}

func handleConn(ctx *gin.Context, conn *websocket.Conn, router *event.Router, l *lobby.Lobby) {
	cl := event.NewClient(conn, "")
	defer l.Disconnect(cl)

	for {
		msg, err := event.Read(ctx, conn)
		if err != nil {
			return
		}
		if err := router.Route(ctx, cl, msg); err != nil {
			return
		}
	}
}

func setupRouter(l *lobby.Lobby) *event.Router {
	router := event.NewRouter()
	router.Handle(event.EventCreate, l.HandleCreate)
	router.Handle(event.EventJoin, l.HandleJoin)
	router.Handle(event.EventMark, l.HandleMark)
	router.Handle(event.EventLeave, l.HandleLeave)
	return router
}
