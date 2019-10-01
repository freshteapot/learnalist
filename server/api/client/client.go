package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/api"
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
	GetAlist(uuid string) (statusCode int, aList alist.Alist, error error)
	PostAlist(body io.Reader) (statusCode int, aList alist.Alist, error error)
	PutAlist(uuid string, body io.Reader) (statusCode int, aList alist.Alist, error error)
	DeleteAlist(uuid string) (statusCode int, error error)
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

func (clientApi *Client) GetAlist(uuid string) (int, *alist.Alist, error) {
	var httpClient = clientApi.getHttpClient()
	url := fmt.Sprintf("%s/alist/%s", clientApi.getServerPath(), uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(clientApi.Config.Username, clientApi.Config.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil, nil
	}

	aList := new(alist.Alist)
	jsonBytes, err := ioutil.ReadAll(resp.Body)
	err = aList.UnmarshalJSON(jsonBytes)
	return resp.StatusCode, aList, err
}

func (clientApi *Client) PutAlist(uuid string, body io.Reader) (int, *alist.Alist, error) {
	var httpClient = clientApi.getHttpClient()
	url := fmt.Sprintf("%s/alist/%s", clientApi.getServerPath(), uuid)
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(clientApi.Config.Username, clientApi.Config.Password)
	req.Header.Set("Content-Type", "javascript")

	resp, err := httpClient.Do(req)

	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	jsonBytes, err := ioutil.ReadAll(resp.Body)
	aList := new(alist.Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	return resp.StatusCode, aList, err
}

func (clientApi *Client) PostAlist(body io.Reader) (int, *alist.Alist, error) {
	var httpClient = clientApi.getHttpClient()
	url := fmt.Sprintf("%s/alist", clientApi.getServerPath())
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(clientApi.Config.Username, clientApi.Config.Password)
	req.Header.Set("Content-Type", "javascript")

	resp, err := httpClient.Do(req)

	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	jsonBytes, err := ioutil.ReadAll(resp.Body)
	aList := new(alist.Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	return resp.StatusCode, aList, err
}

func (clientApi *Client) DeleteAlist(uuid string) (int, error) {
	var httpClient = clientApi.getHttpClient()
	url := fmt.Sprintf("%s/alist/%s", clientApi.getServerPath(), uuid)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(clientApi.Config.Username, clientApi.Config.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
	return resp.StatusCode, err
}
