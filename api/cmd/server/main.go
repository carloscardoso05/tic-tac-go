package main

import (
	"api/event"
	"api/internal/client"
	"api/internal/lobby"
	"api/internal/transport"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	l := lobby.New()
	router := setupRouter(l)

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		conn, err := transport.Accept(c)
		if err != nil {
			return
		}

		handleConn(c, conn, router, l)
		conn.CloseNow()
	})

	r.Run()
}

func handleConn(ctx *gin.Context, conn *websocket.Conn, router *event.Router, l *lobby.Lobby) {
	cl := client.New(conn, "")
	defer l.Disconnect(cl)

	for {
		msg, err := transport.Read(ctx, conn)
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
	router.Handle("create", l.HandleCreate)
	router.Handle("join", l.HandleJoin)
	router.Handle("mark", l.HandleMark)
	router.Handle("leave", l.HandleLeave)
	return router
}
