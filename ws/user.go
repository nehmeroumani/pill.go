package ws

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nehmeroumani/pill.go/clean"
)

type User struct {
	// The websocket connection.
	ws *websocket.Conn
	ID int32
	// Buffered channel of outbound messages.
	Send chan []byte
}

func NewUser(userId int32, ws *websocket.Conn) *User {
	return &User{ws: ws, ID: userId, Send: make(chan []byte, 256)}
}

func (this *User) Write(msgType int, payload []byte) error {
	this.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return this.ws.WriteMessage(msgType, payload)
}

func (this *User) ReadPump() {
	defer func() {
		this.ws.Close()
	}()
	this.ws.SetReadLimit(maxMessageSize)
	this.ws.SetReadDeadline(time.Now().Add(pongWait))
	this.ws.SetPongHandler(func(string) error {
		this.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := this.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	}
}

func (this *User) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		this.ws.Close()
		DefaultHub.Unregister <- this
	}()
	for {
		select {
		case message, ok := <-this.Send:
			if !ok {
				this.Write(websocket.CloseMessage, []byte{})
				return
			}
			this.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := this.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				clean.Error(err)
				return
			}
			w.Write(message)
			if err = w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := this.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
