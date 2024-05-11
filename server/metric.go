package server

import (
	"fmt"
	"github.com/FiveM-Analytics/websocket-server/api"
	"log"
	"time"
)

const (
	interval = time.Second * 10
)

type Metric struct {
	MetricOpts

	Server *WebsocketServer
}

type MetricOpts struct {
	Interval time.Duration
}

func NewMetric(s *WebsocketServer, opts MetricOpts) *Metric {
	return &Metric{
		Server:     s,
		MetricOpts: opts,
	}
}

func (m *Metric) MainLoop() {

	for {
		go m.handleMetrics()

		time.Sleep(m.Interval)
	}
}

type MetricRequest struct {
	Type string `json:"type"`
}

func (m *Metric) handleMetrics() {
	for _, client := range m.Server.Clients {
		if client.Preferences != nil {
			for metric, enabled := range client.Preferences.Analytics {
				if enabled {
					metricRequest := &MetricRequest{
						Type: metric,
					}

					if err := client.Conn.WriteJSON(metricRequest); err != nil {
						log.Println("write err", err)
						continue
					}
				}
			}
		}

	}
}

func (m *Metric) Message(c *Client, payload ServerData) error {
	log.Printf("[%s] recv new message\n", c.Conn.RemoteAddr())
	log.Printf("%+v\n", payload)

	for key, value := range payload.Data {
		sdk := api.NewApi()
		status, err := sdk.SendMetric(c.Id, map[string]any{
			"type": key,
			"data": value,
		})
		if err != nil {
			return err
		}

		if status != 200 {
			return fmt.Errorf("send metric err (%d)\n", status)
		}
	}

	return nil
}
