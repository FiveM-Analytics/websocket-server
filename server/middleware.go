package server

import (
	"errors"
	"fmt"
	"github.com/FiveM-Analytics/websocket-server/api"
)

type MiddlewareFunc func(*Client) error

type Middleware struct {
	Funcs []MiddlewareFunc
}

func NewMiddleware(funcs ...MiddlewareFunc) *Middleware {
	return &Middleware{
		Funcs: funcs,
	}
}

func (m *Middleware) Check(c *Client) error {
	for _, fn := range m.Funcs {
		if err := fn(c); err != nil {
			fmt.Println("dropping client", err)
			return err
		}
	}

	return nil
}

func ServerMiddleware(c *Client) error {
	sdk := api.NewApi()
	// TODO: Somehow store the server info?
	var clientData ClientData
	status, err := sdk.CheckServer(c.EncryptedId, c.Name, &clientData)
	if err != nil {
		return err
	}

	if status == 200 || status == 201 {
		c.refreshed <- clientData
		//c.ClientData = clientData
		//if c.ClientData.Preferences != nil {
		//	// TODO: refresht niet
		//	for key, data := range c.ClientData.Preferences.Analytics {
		//		fmt.Printf("[Middleware] %s: %d\n", key, data.Interval)
		//	}
		//}

		return nil
	}

	return errors.New("Invalid Server")
}
