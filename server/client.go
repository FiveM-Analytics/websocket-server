package server

import (
	"encoding/json"
	"fmt"
	"github.com/FiveM-Analytics/websocket-server/api"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	EncryptedId string
	Name        string
	*websocket.Conn
	Server *WebsocketServer
	ClientData
	send chan interface{}
}

type ClientData struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Preferences *ClientPreferences `json:"preferences"`
}

type ClientPreferences struct {
	Analytics map[string]bool `json:"analytics"`
}

type ClientMessage struct {
	Client  *Client
	Message []byte
}

func NewClient(s *WebsocketServer, id string, name string, conn *websocket.Conn) *Client {
	return &Client{
		EncryptedId: id,
		Name:        name,
		Conn:        conn,
		Server:      s,
		send:        make(chan interface{}),
	}
}

func (c *Client) Refresh() {
	for {
		middleware := NewMiddleware(ServerMiddleware)
		if err := middleware.Check(c); err != nil {
			err = c.Conn.Close()
			if err != nil {
				fmt.Println("close after middleware:", err)
			}
			break
		}

		time.Sleep(1 * time.Minute)
	}
}

func (c *Client) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			log.Println("Websocket connection cannot close", err.Error())
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {

			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			buf, err := json.Marshal(&message)
			if err != nil {
				log.Println(err)
			}

			if _, err = w.Write(buf); err != nil {
				log.Println("write err", err)
			}

			if err = w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println(err)
			}

			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReceiveLoop() {
	sdk := api.NewApi()

	defer func() {
		c.Server.Unregister <- c
		sdk.Disconnect(c.Id)
		err := c.Conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	sdk.Connect(c.Id)
	log.Printf("[%s] Client (%s) connected", c.Conn.RemoteAddr(), c.Id)

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error: %s\n", err)
			}
			break
		}

		log.Printf("[%s] received (%d) bytes from client (%s)", c.Conn.RemoteAddr(), len(message), c.Id)
		c.Server.Dispatch <- &ClientMessage{
			Client:  c,
			Message: message,
		}
	}
}
