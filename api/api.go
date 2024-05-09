package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type ApiOpts struct {
	BaseUrl string
}

type Api struct {
	ApiOpts
}

type CheckServerResponse struct {
}

func NewApi() *Api {
	return &Api{
		ApiOpts: ApiOpts{
			BaseUrl: os.Getenv("API_BASE_URL"),
		},
	}
}

func (api *Api) Connect(id string) (int, error) {
	return api.post(fmt.Sprintf("/servers/connect/%s", id), nil, nil)
}

func (api *Api) Disconnect(id string) (int, error) {
	return api.post(fmt.Sprintf("/servers/disconnect/%s", id), nil, nil)
}

func (api *Api) CheckServer(id string, name string, data any) (int, error) {
	status, err := api.post("/servers/check", data, map[string]any{
		"id":   id,
		"name": name,
	})

	if err != nil {
		return status, err
	}

	return status, nil
}

//func (api *Api) get(url string) {
//	url, _ = strings.CutPrefix(url, "/")
//	resp, err := http.Get(fmt.Sprintf("%s/%s", api.getBaseUrl(), url))
//
//	return nil
//}

func (api *Api) post(url string, respData any, data map[string]any) (int, error) {
	postData, _ := json.Marshal(data)
	buf := bytes.NewBuffer(postData)

	url, _ = strings.CutPrefix(url, "/")
	resp, err := http.Post(fmt.Sprintf("%s/%s", api.getBaseUrl(), url), "application/json", buf)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	log.Printf("(%s) POST (%d)", url, resp.StatusCode)

	if resp.StatusCode == 404 || resp.StatusCode == 500 || resp.StatusCode == 503 {
		return resp.StatusCode, nil
	}

	err = json.NewDecoder(resp.Body).Decode(respData)
	return resp.StatusCode, err
}

func (api *Api) getBaseUrl() string {
	return fmt.Sprintf("%s/api", api.ApiOpts.BaseUrl)
}
