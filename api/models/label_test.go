package models

import (
	"net/http"
	"testing"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestPostUserLabel(t *testing.T) {
	resetDatabase()
	var statusCode int
	var err error

	a := NewUserLabel("label1", "user2")
	statusCode, _ = dal.PostUserLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)
	// Check duplicate entry returns 200.
	statusCode, _ = dal.PostUserLabel(a)
	assert.Equal(t, http.StatusOK, statusCode)

	b := NewUserLabel("label_123456789_123456789_car_boat", "user2")
	statusCode, err = dal.PostUserLabel(b)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, i18n.ValidationWarningLabelToLong, err.Error())

	c := NewUserLabel("", "user2")
	statusCode, err = dal.PostUserLabel(c)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, i18n.ValidationWarningLabelNotEmpty, err.Error())
}

func TestPostAlistLabel(t *testing.T) {
	resetDatabase()
	a := NewAlistLabel("label1", "u:123", "u:456")
	statusCode, _ := dal.PostAlistLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, _ = dal.PostAlistLabel(a)
	assert.Equal(t, http.StatusOK, statusCode)

	b := NewAlistLabel("label_123456789_123456789_car_boat", "u:123", "u:456")
	statusCode, err := dal.PostAlistLabel(b)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, i18n.ValidationWarningLabelToLong, err.Error())
}

// TODO refactor to use sqlite, as this is annoying.
func TestGetUserLabels(t *testing.T) {
	resetDatabase()
	setup := `
INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":[]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','A9046052444752556320','chris');
`
	dal.Db.MustExec(setup)
	alist_uuid := "ada41576-b710-593a-9603-946aaadcb22d"
	user_uuid := "7540fe5f-9847-5473-bdbd-2b20050da0c6"

	// Testing for an empty response
	emptyList := make([]string, 0)
	labels, _ := dal.GetUserLabels(user_uuid)
	assert.Equal(t, emptyList, labels)

	a := NewUserLabel("label1", user_uuid)
	statusCode, _ := dal.PostUserLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)

	labels, _ = dal.GetUserLabels(user_uuid)

	b := NewAlistLabel("label2", user_uuid, alist_uuid)
	statusCode, _ = dal.PostAlistLabel(b)
	assert.Equal(t, http.StatusCreated, statusCode)
	// We add the same one, just to make sure we only get 2 back.
	statusCode, _ = dal.PostAlistLabel(b)
	labels, _ = dal.GetUserLabels(user_uuid)
	assert.Equal(t, 2, len(labels))
}

/*
For setup data.

curl -s -w "%{http_code}\n" -XPOST 127.0.0.1:1234/register -d'
{
    "username":"test1",
    "password":"test"
}
'

curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'test1:test' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v1",
        "labels": [
          "label a",
          "label b"
        ]
    }
}'

*/
func TestRemoveLabelsFromExistingLists(t *testing.T) {
	resetDatabase()
	setup := `
INSERT INTO user VALUES('c3d330fd-73b9-5d3c-840e-3bb59367b5ed','A4187246584872952904','test1');
INSERT INTO alist_kv VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','v1','{"data":[],"info":{"title":"Days of the Week","type":"v1","labels":["label a","label b"]},"uuid":"3c9394eb-7df1-5611-bce1-2bc3a198b2a6"}','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO user_labels VALUES('label a','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO user_labels VALUES('label b','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO alist_labels VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','c3d330fd-73b9-5d3c-840e-3bb59367b5ed','label a');
INSERT INTO alist_labels VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','c3d330fd-73b9-5d3c-840e-3bb59367b5ed','label b');
`
	alist_uuid := "3c9394eb-7df1-5611-bce1-2bc3a198b2a6"
	user_uuid := "c3d330fd-73b9-5d3c-840e-3bb59367b5ed"
	label_remove := "label a"
	label_find := "label b"
	dal.Db.MustExec(setup)
	// Confirm the data is correct with two items for the list.
	alist, _ := dal.GetAlist(alist_uuid)
	assert.Equal(t, 2, len(alist.Info.Labels))
	// Confirm the data is correct with two items.
	labels, _ := dal.GetUserLabels(user_uuid)
	assert.Equal(t, 2, len(labels))

	err := dal.RemoveUserLabel(label_remove, user_uuid)
	assert.NoError(t, err)
	// Confirm the func returns the correct data.
	labels, _ = dal.GetUserLabels(user_uuid)
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, label_find, labels[0])
	// Confirm the list has been updated
	alist, _ = dal.GetAlist(alist_uuid)
	assert.Equal(t, 1, len(alist.Info.Labels))
	assert.Equal(t, label_find, alist.Info.Labels[0])
}
