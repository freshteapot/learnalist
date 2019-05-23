package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
)

const (
	SimpleList    = "v1"
	FromToList    = "v2"
	Concept2      = "v3"
	ContentAndUrl = "v4"
)

var allowedListTypes = []string{
	SimpleList,
	FromToList,
	Concept2,
	ContentAndUrl,
}

// AlistInfo info about the list. Generic to all lists.
type AlistInfo struct {
	Title    string   `json:"title"`
	ListType string   `json:"type"`
	Labels   []string `json:"labels"`
	From     string   `json:"from,omitempty"` // If from is set, we return it, so the 3rd party has context.
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
		aList.Data, err = parseTypeV1(jsonBytes)
		if err != nil {
			err = errors.New(i18n.ValidationErrorListV1)
			return err
		}
	case FromToList:
		aList.Data, err = parseTypeV2(jsonBytes)
		if err != nil {
			err = errors.New(i18n.ValidationErrorListV2)
			return err
		}
	case Concept2:
		aList.Data, err = parseTypeV3(jsonBytes)
		if err != nil {
			err = errors.New(i18n.ValidationErrorListV3)
			return err
		}
		// TODO This is ugly
		*aList = enrichTypeV3(*aList)
	case ContentAndUrl:
		aList.Data, err = parseTypeV4(jsonBytes)
		if err != nil {
			err = errors.New(i18n.ValidationErrorListV4)
			return err
		}
	default:
		err = errors.New("Unsupported list type.")
		return err
	}
	return nil
}
