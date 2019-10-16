package main

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 100 * time.Second
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type member struct {
	name    string
	hub     *hub
	conn    *websocket.Conn
	payload chan string
}

func (m *member) read() {
	defer func() {
		m.conn.Close()
	}()

	m.conn.SetReadLimit(maxMessageSize)
	m.conn.SetReadDeadline(time.Now().Add(pongWait))
	m.conn.SetPongHandler(func(string) error { m.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := m.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// when member is disconnect, we must remove member from hub
				m.hub.removeMember <- m
			}
			break
		}
	}
}

func (m *member) receiveAndWrite() {
	defer func() {
		close(m.payload)
		m.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-m.payload:
			m.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				m.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// sending message to websocket
			if err := m.conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				log.Fatalf("[MEMBER-LISTENER] Error when write message to websocket : %s", err)
			}

			// leave from hub
			m.hub.delete <- m.name
			return
		}
	}
}

func listen(hub *hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	name, _ := c.Params.Get("name")

	// init member
	member := &member{
		name:    name,
		hub:     hub,
		conn:    conn,
		payload: make(chan string),
	}

	// join to hub
	hub.join <- member

	go member.receiveAndWrite()
	go member.read()
}
