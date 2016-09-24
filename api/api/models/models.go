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
	Title    string
	listType string
}

// Alist the outer wrapping of a list.
type Alist struct {
	uuid     string
	listType string
	info     AlistInfo
	data     interface{}
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
		"type":  a.listType,
	})
}

// MarshalJSON convert alist into json
func (a Alist) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"uuid": a.uuid,
		"info": a.info,
		"data": a.data,
	})
}

// GetListsBy Get all alists by uuid
func (db *DB) GetListsBy(uuid string) ([]*Alist, error) {
	// @todo use userid and not return all lists.
	rows, err := db.Query("SELECT uuid FROM alist")
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
		item, err = db.GetAlist(uuid)
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
func (db *DB) GetAlist(uuid string) (*Alist, error) {
	stmt, err := db.Prepare("select uuid, list_type as listType, info, data from alist where uuid = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var sInfo string
	var sData string
	info := new(AlistInfo)
	item := new(Alist)
	err = stmt.QueryRow(uuid).Scan(&item.uuid, &item.listType, &sInfo, &sData)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(sInfo), &info)

	//Could maybe use this if I want to get fancy
	if item.listType == "v1" {
		dataV1 := new(AlistTypeV1)
		json.Unmarshal([]byte(sData), &dataV1)
		item.data = *dataV1
	} else if item.listType == "v2" {
		fmt.Println(sData)
		dataV2 := new(AlistTypeV2)
		json.Unmarshal([]byte(sData), &dataV2)
		item.data = *dataV2
	} else {
		json.Unmarshal([]byte(sData), &item.data)
	}

	info.listType = item.listType
	item.info = *info

	return item, nil
}

// @todo
// PostAlist Process user data and store as new in the db.
func (db *DB) PostAlist(interface{}) (*Alist, error) {
	uuid := "123"
	return db.GetAlist(uuid)
}

// @todo
// UpdateAlist Process user data and store in db as an update.
func (db *DB) UpdateAlist(interface{}) (*Alist, error) {
	uuid := "123"
	return db.GetAlist(uuid)
}

// CreateDBStructure Create the database tables
func (db *DB) CreateDBStructure() {
	query := "create table alist (uuid CHARACTER(36)  not null primary key, list_type CHARACTER(3), info text, data text);"
	_, err := db.Exec(query)
	if err != nil {
		// table alist already exists
		return
	}
	// Add some test data
	query = `
INSERT INTO alist values ('230bf9f8-592b-55c1-8f72-9ea32fbdcdc4', 'v1', '{"title":"I am a list"}', '["a","b"]');
INSERT INTO alist values ('efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9', 'v2', '{"title":"I am a list with items"}', '{"car":"bil", "water": "vann"}');
`
	db.Exec(query)
}
