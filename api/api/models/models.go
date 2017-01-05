package models

import (
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

// AlistItemTypeV2 Item in  AlistTypeV2
type AlistItemTypeV2 struct {
	From string
	To   string
}

// AlistTypeV2 list type v2
type AlistTypeV2 struct {
	Items []AlistItemTypeV2
}

// AlistTypeV1 list type v1
type AlistTypeV1 []string

// AlistInfo info about the list. Generic to all lists.
type AlistInfo struct {
	Title    string `json:"title"`
	ListType string `json:"type"`
}

type InputAlist struct {
	*Alist
}

// Alist the outer wrapping of a list.
type Alist struct {
	Uuid     string `json:"uuid"`
	ListType string
	Info     AlistInfo   `json:"info"`
	Data     interface{} `json:"data"`
}

// UnmarshalJSON convert list type v2 from json
func (items *AlistTypeV2) UnmarshalJSON(data []byte) error {
	var stuff map[string]string
	err := json.Unmarshal(data, &stuff)
	if err != nil {
		return err
	}
	for key, value := range stuff {
		item := AlistItemTypeV2{From: key, To: value}
		items.Items = append(items.Items, item)
	}
	return nil

}

// MarshalJSON convert list type v2 into json
func (data AlistTypeV2) MarshalJSON() ([]byte, error) {
	response := make(map[string]string)
	for _, v := range data.Items {
		response[v.From] = v.To
	}
	return json.Marshal(response)
}

// MarshalJSON convert list type v2 item into json
func (data AlistItemTypeV2) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		data.From: data.To,
	})
}

// MarshalJSON convert list info into json
func (a AlistInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"title": a.Title,
		"type":  a.ListType,
	})
}

// MarshalJSON convert alist into json
func (a Alist) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"uuid": a.Uuid,
		"info": a.Info,
		"data": a.Data,
	})
}

// GetListsBy Get all alists by uuid
func (dal *DAL) GetListsBy(uuid string) ([]*Alist, error) {
	// @todo use userid and not return all lists.
	rows, err := dal.Db.Query("SELECT uuid FROM alist")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Alist

	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		var item *Alist
		item, err = dal.GetAlist(uuid)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetAlist Get alist
func (dal *DAL) GetAlist(uuid string) (*Alist, error) {
	stmt, err := dal.Db.Prepare("select uuid, list_type as listType, info, data from alist where uuid = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	fmt.Println("After")
	var sInfo string
	var sData string
	info := new(AlistInfo)
	item := new(Alist)
	err = stmt.QueryRow(uuid).Scan(&item.Uuid, &item.ListType, &sInfo, &sData)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(sInfo), &info)

	//Could maybe use this if I want to get fancy
	if item.ListType == "v1" {
		dataV1 := new(AlistTypeV1)
		json.Unmarshal([]byte(sData), &dataV1)
		item.Data = *dataV1
	} else if item.ListType == "v2" {
		fmt.Println(sData)
		dataV2 := new(AlistTypeV2)
		json.Unmarshal([]byte(sData), &dataV2)
		item.Data = *dataV2
	} else {
		json.Unmarshal([]byte(sData), &item.Data)
	}

	info.ListType = item.ListType
	item.Info = *info

	return item, nil
}

// @todo
// PostAlist Process user data and store as new in the db.
func (dal *DAL) PostAlist(interface{}) (*Alist, error) {
	uuid := "123"
	return dal.GetAlist(uuid)
}

// @todo
// UpdateAlist Process user data and store in db as an update.
func (dal *DAL) UpdateAlist(interface{}) (*Alist, error) {
	uuid := "123"
	return dal.GetAlist(uuid)
}

// CreateDBStructure Create the database tables
func (dal *DAL) CreateDBStructure() {
	query := "create table alist (uuid CHARACTER(36)  not null primary key, list_type CHARACTER(3), info text, data text);"
	_, err := dal.Db.Exec(query)
	if err != nil {
		// table alist already exists
		return
	}
	// Add some test data
	query = `
INSERT INTO alist values ('230bf9f8-592b-55c1-8f72-9ea32fbdcdc4', 'v1', '{"title":"I am a list"}', '["a","b"]');
INSERT INTO alist values ('efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9', 'v2', '{"title":"I am a list with items"}', '{"car":"bil", "water": "vann"}');
`
	dal.Db.Exec(query)
}
