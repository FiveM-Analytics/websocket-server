package main

import (
	"flag"
	"github.com/FiveM-Analytics/websocket-server/server"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	MB = 1024
)

func main() {
	flag.Parse()

	// Initialize the environment file
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		panic(err)
	}

	serverOpts := server.WebsocketServerOpts{
		Host: os.Getenv("APP_HOST"),
		Port: port,
		Upgrade: websocket.Upgrader{
			ReadBufferSize:  MB * 1,
			WriteBufferSize: MB * 1,

			CheckOrigin: func(r *http.Request) bool {
				// For development only !!!
				env := os.Getenv("APP_ENV")
				switch env {
				case "local", "dev":
					return true

				case "staging":
					allowedOrigin := "https://**"
					origin := r.Header.Get("Origin")
					return origin == allowedOrigin

				case "production":
					allowedOrigin := "https://**"
					origin := r.Header.Get("Origin")
					return origin == allowedOrigin
				}
				return false
			},
		},
	}

	metricOpts := server.MetricOpts{
		Interval: time.Second * 15,
	}

	s := server.NewWebsocketServer(serverOpts, metricOpts)

	if err = s.Run(); err != nil {
		panic(err)
	}
}
