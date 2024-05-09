package server

import (
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

func (m *Metric) Message(c *Client, payload ServerData) {

}
