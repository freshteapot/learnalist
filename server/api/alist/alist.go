package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
)

const (
	SimpleList       = "v1"
	FromToList       = "v2"
	Concept2         = "v3"
	ContentAndUrl    = "v4"
	InteractEnabled  = 1
	InteractDisabled = 0
)

var ValidInteract = []int{InteractEnabled, InteractDisabled}

var allowedListTypes = []string{
	SimpleList,
	FromToList,
	Concept2,
	ContentAndUrl,
}

type ShortInfo struct {
	UUID  string `db:"uuid" json:"uuid"`
	Title string `db:"title" json:"title"`
}

// AlistInfo info about the list. Generic to all lists.
type AlistInfo struct {
	Title      string    `json:"title"`
	ListType   string    `json:"type"`
	Labels     []string  `json:"labels"`
	From       string    `json:"from,omitempty"` // If from is set, we return it, so the 3rd party has context.
	Interact   *Interact `json:"interact,omitempty"`
	SharedWith string    `json:"shared_with"`
}

type Interact struct {
	Slideshow   int `json:"slideshow"`
	TotalRecall int `json:"totalrecall"`
}

type InputAlist struct {
	*Alist
}

// Alist the outer wrapping of a list.
type Alist struct {
	Uuid string `json:"uuid"`
	User uuid.User
	Info AlistInfo   `json:"info"`
	Data interface{} `json:"data"`
}

type AlistTypeMarshalJSON interface {
	ParseInfo(info AlistInfo) (AlistInfo, error)
	ParseData([]byte) (interface{}, error)
	Enrich(Alist) Alist
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

	if !utils.StringArrayContains(allowedListTypes, aList.Info.ListType) {
		err = errors.New("Unsupported list type.")
		return err
	}

	if raw["data"] == nil {
		err = errors.New("Failed to pass list. Data is missing.")
		return err
	}

	var mapper AlistTypeMarshalJSON
	switch aList.Info.ListType {
	case SimpleList:
		mapper = NewMapToV1()
	case FromToList:
		mapper = NewMapToV2()
		break
	case Concept2:
		mapper = NewMapToV3()
		break
	case ContentAndUrl:
		mapper = NewMapToV4()
		break
	}

	aList.Info, _ = mapper.ParseInfo(aList.Info)

	jsonBytes, _ = json.Marshal(raw["data"])
	aList.Data, err = mapper.ParseData(jsonBytes)
	if err != nil {
		return err
	}
	*aList = mapper.Enrich(*aList)
	return nil
}
