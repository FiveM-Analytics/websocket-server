package server

import (
	"encoding/json"
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

type MetricRequest struct {
	Type string `json:"type"`
}

func (m *Metric) Client(client *Client) {
	if client.Preferences != nil {
		for metric, analytics := range client.Preferences.Analytics {
			go m.handleMetricLoop(client, metric, analytics)
		}
	}
}

func (m *Metric) handleMetricLoop(client *Client, metric string, analytics *ClientAnalytics) {
	for {
		if analytics.Enabled {
			metricRequest := &MetricRequest{
				Type: metric,
			}

			client.send <- metricRequest
		} else {
			time.Sleep(m.Interval)
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%dms", analytics.Interval))
		if err != nil {
			time.Sleep(m.Interval)
		} else {
			fmt.Printf("[Metric] %s: %d\n", metric, analytics.Interval)
			time.Sleep(duration)
		}
	}
}

type MetricMessage struct {
	Data map[string]any `json:"data"`
}

func (m *Metric) Message(c *Client, message []byte) error {
	log.Printf("[%s] recv new message\n", c.Conn.RemoteAddr())

	var payload MetricMessage
	if err := json.Unmarshal(message, &payload); err != nil {
		log.Printf("unmarshal err: %v\n", err)
	}

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
