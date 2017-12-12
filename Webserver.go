package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan *Timeular
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

const (
	// Time allowed to write the file to the server.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the server.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func RunWebserver(hub *Hub) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWebSocket(hub, w, r)
	})
	http.Handle("/", http.FileServer(http.Dir("web/")))

	log.Printf("Webserver started on port 6677")
	log.Fatal(http.ListenAndServe(":6677", nil))
}

func serveWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	client := &Client{hub: hub, conn: conn, send: make(chan *Timeular)}
	client.hub.register <- client

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go client.writer()
	client.reader()
}

func (c *Client) reader() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Println("Pong")
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		messageType, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType == websocket.TextMessage {
			log.Printf("Message (type %d): %s", messageType, data)
		} else {
			log.Printf("Read type %d", messageType)
		}
	}
}

func (c *Client) writer() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case state := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			message, err := json.Marshal(state)
			if err != nil {
				log.Printf("Could not marshal Timeular state")
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error: %s", err)
				return
			}
		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Error: %s", err)
				return
			}
		}
	}
}
