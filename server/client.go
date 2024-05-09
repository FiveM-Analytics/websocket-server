package server

import (
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
	ClientData
}

type ClientData struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Preferences *ClientPreferences `json:"preferences"`
}

type ClientPreferences struct {
	Analytics map[string]bool `json:"analytics"`
}

type ServerData struct {
	Data map[string]any `json:"data"`
}

func NewClient(id string, name string, conn *websocket.Conn) *Client {
	return &Client{
		EncryptedId: id,
		Name:        name,
		Conn:        conn,
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

func (c *Client) ReceiveLoop() {
	sdk := api.NewApi()
	sdk.Connect(c.Id)
	log.Printf("[%s] Client (%s) connected", c.Conn.RemoteAddr(), c.Id)

	defer func() {
		sdk.Disconnect(c.Id)
		err := c.Conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	//c.Conn.SetReadLimit(maxMessageSize)
	//_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	//c.Conn.SetPongHandler(func(string) error {
	//	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	//	return nil
	//})

	for {
		var data ServerData
		if err := c.Conn.ReadJSON(&data); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				log.Printf("[%s] Client (%s) disconnected: %s", c.Conn.RemoteAddr(), c.Id, err)
				break
			}

			log.Println("read error", err)
			continue
		}

	}
}
