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

	Clients []*Client
}

func NewWebsocketServer(serverOpts WebsocketServerOpts, metricOpts MetricOpts) *WebsocketServer {
	s := &WebsocketServer{
		WebsocketServerOpts: serverOpts,
	}

	s.Metric = NewMetric(s, metricOpts)

	return s
}

func (s *WebsocketServer) routes() {
	http.HandleFunc("/metrics", s.serve)
}

func (s *WebsocketServer) Run() error {
	s.routes()

	go s.Metric.MainLoop()

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

	client := NewClient(id, name, conn)

	middleware := NewMiddleware(ServerMiddleware)
	if err = middleware.Check(client); err != nil {
		err = conn.Close()
		if err != nil {
			fmt.Println("close after middleware:", err)
		}
		return
	}

	s.Clients = append(s.Clients, client)

	go client.Refresh()
	go client.ReceiveLoop(s)
}
