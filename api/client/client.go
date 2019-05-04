package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/api/api"
)

type Config struct {
	Server   string
	Username string
	Password string
}

type Client struct {
	Config Config
}

func (clientApi *Client) getServerPath() string {
	return strings.TrimSuffix(clientApi.Config.Server, "/")
}

func (clientApi *Client) getHttpClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}
	return netClient
}

type LearnalistAPI interface {
	GetRoot() (api.HttpResponseMessage, error)
	GetVersion() (api.HttpGetVersionResponse, error)
}

func (clientApi *Client) GetRoot() (api.HttpResponseMessage, error) {
	var httpResponse api.HttpResponseMessage
	netClient := clientApi.getHttpClient()
	url := fmt.Sprintf("%s/", clientApi.getServerPath())
	response, err := netClient.Get(url)

	if err != nil {
		return httpResponse, err
	}

	if response.StatusCode != http.StatusOK {
		return httpResponse, errors.New("Failed.")
	}

	buf, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &httpResponse)
	return httpResponse, nil
}

func (clientApi *Client) GetVersion() (api.HttpGetVersionResponse, error) {
	var httpResponse api.HttpGetVersionResponse
	netClient := clientApi.getHttpClient()
	url := fmt.Sprintf("%s/version", clientApi.getServerPath())
	response, err := netClient.Get(url)

	if err != nil {
		return httpResponse, err
	}

	if response.StatusCode != http.StatusOK {
		return httpResponse, errors.New("Failed.")
	}

	buf, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &httpResponse)
	return httpResponse, nil
}
