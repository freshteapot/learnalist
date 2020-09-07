package models_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/label"
)

func (suite *ModelSuite) TestPostUserLabel() {
	var statusCode int
	var err error

	a := label.NewUserLabel("label1", "user2")
	statusCode, _ = dal.Labels().PostUserLabel(a)
	suite.Equal(http.StatusCreated, statusCode)
	// Check duplicate entry returns 200.
	statusCode, _ = dal.Labels().PostUserLabel(a)
	suite.Equal(http.StatusOK, statusCode)

	b := label.NewUserLabel("label_123456789_123456789_car_boat", "user2")
	statusCode, err = dal.Labels().PostUserLabel(b)
	suite.Equal(http.StatusBadRequest, statusCode)
	suite.Equal(i18n.ValidationWarningLabelToLong, err.Error())

	c := label.NewUserLabel("", "user2")
	statusCode, err = dal.Labels().PostUserLabel(c)
	suite.Equal(http.StatusBadRequest, statusCode)
	suite.Equal(i18n.ValidationWarningLabelNotEmpty, err.Error())
}

func (suite *ModelSuite) TestPostAlistLabel() {
	a := label.NewAlistLabel("label1", "u:123", "u:456")
	statusCode, _ := dal.Labels().PostAlistLabel(a)
	suite.Equal(http.StatusCreated, statusCode)

	statusCode, _ = dal.Labels().PostAlistLabel(a)
	suite.Equal(http.StatusOK, statusCode)

	b := label.NewAlistLabel("label_123456789_123456789_car_boat", "u:123", "u:456")
	statusCode, err := dal.Labels().PostAlistLabel(b)
	suite.Equal(http.StatusBadRequest, statusCode)
	suite.Equal(i18n.ValidationWarningLabelToLong, err.Error())
}

func (suite *ModelSuite) TestGetUserLabels() {
	setup := `
INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":[]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
`
	dal.Db.MustExec(setup)
	alist_uuid := "ada41576-b710-593a-9603-946aaadcb22d"
	user_uuid := "7540fe5f-9847-5473-bdbd-2b20050da0c6"

	// Testing for an empty response
	emptyList := make([]string, 0)
	labels, _ := dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(emptyList, labels)

	a := label.NewUserLabel("label1", user_uuid)
	statusCode, _ := dal.Labels().PostUserLabel(a)
	suite.Equal(http.StatusCreated, statusCode)

	labels, _ = dal.Labels().GetUserLabels(user_uuid)

	b := label.NewAlistLabel("label2", user_uuid, alist_uuid)
	statusCode, _ = dal.Labels().PostAlistLabel(b)
	suite.Equal(http.StatusCreated, statusCode)
	// We add the same one, just to make sure we only get 2 back.
	statusCode, _ = dal.Labels().PostAlistLabel(b)
	labels, _ = dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(2, len(labels))
}

/*
For setup data.

curl -s -w "%{http_code}\n" -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'

curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'test1:test' -d'
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
func (suite *ModelSuite) TestRemoveLabelsFromExistingLists() {
	setup := `
INSERT INTO user VALUES('c3d330fd-73b9-5d3c-840e-3bb59367b5ed','4187246584872952904','test1');
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
	suite.Equal(2, len(alist.Info.Labels))
	// Confirm the data is correct with two items.
	labels, _ := dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(2, len(labels))

	err := dal.RemoveUserLabel(label_remove, user_uuid)

	suite.NoError(err)
	// Confirm the func returns the correct data.
	labels, _ = dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(1, len(labels))
	suite.Equal(label_find, labels[0])
	// Confirm the list has been updated
	alist, _ = dal.GetAlist(alist_uuid)
	suite.Equal(1, len(alist.Info.Labels))
	suite.Equal(label_find, alist.Info.Labels[0])
}
