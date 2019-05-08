package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/api/uuid"
)

const (
	SimpleList = "v1"
	FromToList = "v2"
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

// MarshalJSON convert alist into json
func (a Alist) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"uuid": a.Uuid,
		"info": a.Info,
		"data": a.Data,
	})
}

// UnmarshalJSON convert list type v2 from json
func (aList *Alist) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	var err error
	var jsonBytes []byte

	err = json.Unmarshal(data, &raw)
	if err != nil {
		err = errors.New("Failed to parse list.")
		return err
	}

	if raw["uuid"] != nil {
		aList.Uuid = raw["uuid"].(string)
	}

	if raw["info"] == nil {
		err = errors.New("Failed to pass list. Info is missing.")
		return err
	}

	jsonBytes, _ = json.Marshal(raw["info"])
	aList.Info, err = parseAlistInfo(jsonBytes)
	if err != nil {
		err = errors.New("Failed to pass list. Something wrong with info object.")
		return err
	}

	if raw["data"] == nil {
		err = errors.New("Failed to pass list. Data is missing.")
		return err
	}

	jsonBytes, _ = json.Marshal(raw["data"])
	switch aList.Info.ListType {
	case SimpleList:
		aList.Data, err = parseAlistTypeV1(jsonBytes)
		if err != nil {
			err = errors.New("Failed to pass list type v1.")
			return err
		}
	case FromToList:
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
