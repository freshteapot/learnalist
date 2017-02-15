package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/api/uuid"
)

// AlistItemTypeV2 Item in  AlistTypeV2
type AlistItemTypeV2 struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type AlistTypeV2 []AlistItemTypeV2

// AlistTypeV1 list type v1
type AlistTypeV1 []string

// AlistInfo info about the list. Generic to all lists.
type AlistInfo struct {
	Title    string `json:"title"`
	ListType string `json:"type"`
	From     string `json:"from,omitempty"` // If from is set, we return it, so the 3rd party has context.
}

type InputAlist struct {
	*Alist
}

// Alist the outer wrapping of a list.
type Alist struct {
	Uuid     string `json:"uuid"`
	User     uuid.User
	ListType string
	Info     AlistInfo   `json:"info"`
	Data     interface{} `json:"data"`
}

// UnmarshalJSON convert list type v2 from json
func (aList *Alist) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	var err error
	var jsonBytes []byte

	err = json.Unmarshal(data, &raw)
	if err != nil {
		err = errors.New("Failed to pass list.")
		return err
	}

	if raw["uuid"] != nil {
		aList.Uuid = raw["uuid"].(string)
	}

	jsonBytes, err = json.Marshal(raw["info"])
	if err != nil {
		err = errors.New("Failed to pass list.")
		return err
	}

	aList.Info, err = parseAlistInfo(jsonBytes)
	if err != nil {
		err = errors.New("Failed to pass list.")
		return err
	}

	jsonBytes, err = json.Marshal(raw["data"])
	if err != nil {
		err = errors.New("Failed to pass list data.")
		return err
	}

	switch aList.Info.ListType {
	case "v1":
		aList.Data, err = parseAlistTypeV1(jsonBytes)
		if err != nil {
			err = errors.New("Failed to pass list type v1.")
			return err
		}
	case "v2":
		aList.Data, err = parseAlistTypeV2(jsonBytes)
		if err != nil {
			err = errors.New("Failed to pass list type v2.")
			return err
		}
	default:
		err = errors.New("Unsupported list type.")
		return err
	}
	return nil
}

// MarshalJSON convert alist into json
func (a Alist) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"uuid": a.Uuid,
		"info": a.Info,
		"data": a.Data,
	})
}

func parseAlistInfo(jsonBytes []byte) (AlistInfo, error) {
	listInfo := new(AlistInfo)
	err := json.Unmarshal(jsonBytes, &listInfo)
	return *listInfo, err
}
func parseAlistTypeV1(jsonBytes []byte) (AlistTypeV1, error) {
	listData := new(AlistTypeV1)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func parseAlistTypeV2(jsonBytes []byte) (AlistTypeV2, error) {
	listData := new(AlistTypeV2)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}
