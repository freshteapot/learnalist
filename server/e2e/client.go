package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
)

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

type RegisterResponse struct {
	Username  string `json:"username"`
	Uuid      string `json:"uuid"`
	BasicAuth string
}

type AlistUuidResponse struct {
	Uuid string `json:"uuid"`
}

type Client struct {
	server     string
	httpClient *http.Client
}

func NewClient(_server string) Client {
	timeout := 1 * time.Second
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout,
	}
	var netClient = &http.Client{
		Timeout:   timeout,
		Transport: netTransport,
	}

	return Client{
		server:     _server,
		httpClient: netClient,
	}
}

func (c Client) getServerURL() string {
	return c.server
}

func (c Client) doRequest(req *http.Request, dest interface{}, want ...int) (int, api.HttpResponseMessage) {
	var errorMessage api.HttpResponseMessage
	resp, _ := c.httpClient.Do(req)

	defer resp.Body.Close()

	if !utils.IntArrayContains(want, resp.StatusCode) {
		json.NewDecoder(resp.Body).Decode(&errorMessage)
		return resp.StatusCode, errorMessage
	}

	json.NewDecoder(resp.Body).Decode(dest)
	return resp.StatusCode, errorMessage
}

func getBasicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c Client) Register(username string, password string) RegisterResponse {
	fmt.Println("Registering user via Register")
	body := strings.NewReader(fmt.Sprintf(`
{
    "username":"%s",
    "password":"%s"
}
`, username, password))

	url := fmt.Sprintf("%s/api/v1/user/register", c.getServerURL())
	req, err := http.NewRequest("POST", url, body)
	req = req.WithContext(context.Background())
	if err != nil {
		// handle err
		fmt.Println("Failed NewRequest")
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response := RegisterResponse{}
	wanted := []int{http.StatusCreated, http.StatusOK}
	statusCode, _ := c.doRequest(req, &response, wanted...)
	if !utils.IntArrayContains(wanted, statusCode) {
		fmt.Println("Failed to register correctly")
		panic(statusCode)
	}
	response.BasicAuth = getBasicAuth(username, password)
	return response
}

func (c Client) DeleteUser(credentials api.HttpLoginResponse) (statusCode int, response api.HttpResponseMessage) {
	url := fmt.Sprintf("%s/api/v1/user/%s", c.getServerURL(), credentials.UserUUID)
	req, err := http.NewRequest("DELETE", url, nil)
	req = req.WithContext(context.Background())
	if err != nil {
		// handle err
		fmt.Println("Failed NewRequest")
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	statusCode, _ = c.doRequest(req, &response, http.StatusOK)
	return statusCode, response
}

func (c Client) PostListV1(userInfo RegisterResponse, input string) (alist.Alist, error) {
	//fmt.Println("Posting a list via PostListV1")
	var response alist.Alist
	resp, err := c.RawPostListV1(userInfo, input)
	if err != nil {
		// handle err
		panic(err)
		return response, nil
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &response)
	return response, nil
}

func (c Client) PutListV1(userInfo RegisterResponse, uuid string, input string) (alist.Alist, error) {
	fmt.Println("Updating a list via PutListV1")
	var response alist.Alist
	resp, err := c.RawPutListV1(userInfo, uuid, input)

	if err != nil {
		// handle err
		return response, nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &response)
	return response, nil
}

func (c Client) SetListShareV1(userInfo RegisterResponse, alistUUID string, action string) api.HttpResponseMessage {
	body := strings.NewReader(fmt.Sprintf(`{
  "alist_uuid": "%s",
  "action": "%s"
}`, alistUUID, action))
	url := fmt.Sprintf("%s/api/v1/share/alist", c.getServerURL())

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		// handle err
		panic(err)
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Basic "+userInfo.BasicAuth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// handle err
		panic(err)
	}
	defer resp.Body.Close()
	var response api.HttpResponseMessage
	data, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &response)
	if err != nil {
		// handle err
		panic(err)
	}
	return response
}

func (c Client) GetListByUUIDV1(userInfo RegisterResponse, uuid string) api.HttpResponse {
	url := fmt.Sprintf("%s/api/v1/alist/%s", c.getServerURL(), uuid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle err

		panic(err)
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Basic "+userInfo.BasicAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// handle err
		fmt.Println("here")
		panic(err)
	}

	defer resp.Body.Close()

	var response api.HttpResponse
	response.StatusCode = resp.StatusCode
	data, err := ioutil.ReadAll(resp.Body)
	response.Body = data
	return response
}

func (c Client) PostLabelV1(userInfo RegisterResponse, label string) ([]string, error) {
	fmt.Println("Posting a list via PostLabelV1")
	var response []string

	resp, err := c.RawPostLabelV1(userInfo, label)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (c Client) GetLabelsByMeV1(userInfo RegisterResponse) ([]string, error) {
	fmt.Println("GET  labels via GetLabelsByMeV1")
	var response []string

	resp, err := c.RawGetLabelsByMeV1(userInfo)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (c Client) GetListsByMe(userInfo RegisterResponse, labels string, listType string) ([]*alist.Alist, error) {
	var response []*alist.Alist
	resp, err := c.RawGetListsByMe(userInfo, labels, listType)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (c Client) GetAlistHtml(userInfo RegisterResponse, uuid string) (api.HttpResponse, error) {
	var response api.HttpResponse
	var err error
	url := fmt.Sprintf("%s/alist/%s.html", c.getServerURL(), uuid)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// handle err
		return response, err
	}

	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Basic "+userInfo.BasicAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// handle err
		return response, err
	}

	defer resp.Body.Close()

	response.StatusCode = resp.StatusCode
	data, err := ioutil.ReadAll(resp.Body)
	response.Body = data
	return response, err
}

func (c Client) ShareReadAcessV1(userInfo RegisterResponse, alistUUID string, userUUID string, action string) (api.HttpResponse, error) {
	var response api.HttpResponse
	var err error
	url := fmt.Sprintf("%s/api/v1/share/readaccess", c.getServerURL())

	inputAccess := &api.HttpShareListWithUserInput{
		UserUUID:  userUUID,
		AlistUUID: alistUUID,
		Action:    action,
	}

	b, _ := json.Marshal(inputAccess)
	body := strings.NewReader(string(b))
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		// handle err
		return response, err
	}

	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Basic "+userInfo.BasicAuth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// handle err
		return response, err
	}

	defer resp.Body.Close()

	response.StatusCode = resp.StatusCode
	data, err := ioutil.ReadAll(resp.Body)
	response.Body = data
	return response, err
}
