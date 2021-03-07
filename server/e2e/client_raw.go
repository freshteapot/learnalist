package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

func (c Client) RawRequest(request *http.Request) (response *http.Response, err error) {
	request = request.WithContext(context.Background())
	response, err = c.httpClient.Do(request)
	return response, err
}

func (c Client) RawLogin(username string, password string) (*http.Response, error) {
	body := strings.NewReader(fmt.Sprintf(`
{
    "username":"%s",
    "password":"%s"
}
`, username, password))

	url := fmt.Sprintf("%s/api/v1/user/login", c.getServerURL())
	req, err := http.NewRequest("POST", url, body)
	req = req.WithContext(context.Background())
	if err != nil {
		// handle err
		fmt.Println("Failed NewRequest")
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func (c Client) RawPostListV1(credentials openapi.HttpUserLoginResponse, input string) (*http.Response, error) {
	var response *http.Response
	body := strings.NewReader(input)
	url := fmt.Sprintf("%s/api/v1/alist", c.getServerURL())
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawPutListV1(credentials openapi.HttpUserLoginResponse, uuid string, input string) (*http.Response, error) {
	fmt.Println("Updating a list via RawPutListV1")
	var response *http.Response
	body := strings.NewReader(input)
	url := fmt.Sprintf("%s/api/v1/alist/%s", c.getServerURL(), uuid)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawDeleteListV1(credentials openapi.HttpUserLoginResponse, uuid string) (*http.Response, error) {
	fmt.Println("Deleting a list via RawDeleteListV1")
	var response *http.Response
	url := fmt.Sprintf("%s/api/v1/alist/%s", c.getServerURL(), uuid)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawPostLabelV1(credentials openapi.HttpUserLoginResponse, label string) (*http.Response, error) {
	fmt.Println("Posting a list via RawPostLabelV1")
	input := api.HTTPLabelInput{
		Label: label,
	}
	var response *http.Response
	b, _ := json.Marshal(input)
	body := strings.NewReader(string(b))
	url := fmt.Sprintf("%s/api/v1/labels", c.getServerURL())
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawGetLabelsByMeV1(credentials openapi.HttpUserLoginResponse) (*http.Response, error) {
	var response *http.Response
	fmt.Println("GET  labels via RawGetLabelsByMeV1")
	url := fmt.Sprintf("%s/api/v1/labels/by/me", c.getServerURL())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawDeleteLabelV1(credentials openapi.HttpUserLoginResponse, label string) (*http.Response, error) {
	fmt.Println("Posting a list via RawDeleteLabelV1")
	var response *http.Response
	url := fmt.Sprintf("%s/api/v1/labels/%s", c.getServerURL(), label)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawGetListsByMe(credentials openapi.HttpUserLoginResponse, labels string, listType string) (*http.Response, error) {
	var response *http.Response
	uri := fmt.Sprintf("%s/api/v1/alist/by/me", c.getServerURL())
	if labels != "" {
		uri = fmt.Sprintf("%s?labels=%s", uri, labels)
	}
	if listType != "" {
		uri = fmt.Sprintf("%s?list_type=%s", uri, listType)
	}
	if listType != "" && labels != "" {
		uri = fmt.Sprintf("%s?labels=%s&list_type=%s", uri, labels, listType)
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c Client) RawV1(credentials openapi.HttpUserLoginResponse, method string, uri string, input string) (*http.Response, error) {
	var response *http.Response

	url := fmt.Sprintf("%s/%s", c.getServerURL(), strings.TrimPrefix(uri, "/"))
	body := strings.NewReader(input)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		// handle err
		return response, nil
	}
	req = req.WithContext(context.Background())
	req.Header.Set("Authorization", "Bearer "+credentials.Token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
