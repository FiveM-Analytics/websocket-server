package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

type WebsocketServerOpts struct {
	Host    string
	Port    int
	Upgrade websocket.Upgrader
}

type WebsocketServer struct {
	WebsocketServerOpts
	Metric *Metric

	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Dispatch   chan *ClientMessage

	quitch chan struct{}
}

func NewWebsocketServer(serverOpts WebsocketServerOpts, metricOpts MetricOpts) *WebsocketServer {
	s := &WebsocketServer{
		WebsocketServerOpts: serverOpts,
		Clients:             make(map[*Client]bool),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		Dispatch:            make(chan *ClientMessage),
	}

	s.Metric = NewMetric(s, metricOpts)

	return s
}

func (s *WebsocketServer) routes() {
	http.HandleFunc("/metrics", s.serve)
}

func (s *WebsocketServer) Listen() {

	for {
		select {
		case <-s.quitch:
			log.Println("Gracefully shutting down...")
			break
		case client := <-s.Register:
			log.Printf("[%s] Client (%s) connected\n", client.Conn.RemoteAddr(), client.Id)
			s.Clients[client] = true
		case client := <-s.Unregister:
			if _, ok := s.Clients[client]; ok {
				log.Printf("[%s] Client (%s) disconnected\n", client.Conn.RemoteAddr(), client.Id)
				delete(s.Clients, client)
			}
		case message := <-s.Dispatch:
			log.Printf("[%s] Client send (%d) bytes message\n", message.Client.Conn.RemoteAddr(), len(message.Message))

			if err := s.Metric.Message(message.Client, message.Message); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *WebsocketServer) Run() error {
	s.routes()

	go s.Metric.MainLoop()
	go s.Listen()

	sslCertPath := os.Getenv("SSL_CERT_PATH")
	sslKeyPath := os.Getenv("SSL_KEY_PATH")
	if sslCertPath == "" || sslKeyPath == "" {
		log.Printf("Server listening at ws://%s:%d\n", s.Host, s.Port)

		err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
		if err != nil {
			fmt.Println("ListenAndServer:", err)
			return err
		}
	} else {
		log.Printf("Server listening at wss://%s:%d\n", s.Host, s.Port)

		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", s.Port), sslCertPath, sslKeyPath, nil)
		if err != nil {
			fmt.Println("ListenAndServer:", err)
			return err
		}
	}

	return nil
}

func (s *WebsocketServer) serve(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
	}

	id := r.FormValue("id")
	name := r.FormValue("name")

	if id == "" || name == "" {
		fmt.Println("tried to connect but missing id or name")
		err = conn.Close()
		if err != nil {
			fmt.Println("Close:", err)
		}
		return
	}

	client := NewClient(s, id, name, conn)

	middleware := NewMiddleware(ServerMiddleware)
	if err = middleware.Check(client); err != nil {
		err = conn.Close()
		if err != nil {
			fmt.Println("close after middleware:", err)
		}
		return
	}

	s.Register <- client
	go client.Refresh()
	go client.ReceiveLoop()
	go client.WriteLoop()
}
