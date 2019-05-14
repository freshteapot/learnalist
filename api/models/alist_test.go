package models

import (
	"testing"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestSaveAlist(t *testing.T) {
	resetDatabase()
	setup := `
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
`
	dal.Db.MustExec(setup)
	a := `
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}
`
	user := uuid.NewUser()
	user.Uuid = "7540fe5f-9847-5473-bdbd-2b20050da0c6"
	playList := uuid.NewPlaylist(&user)
	alist_uuid := playList.Uuid

	aList := new(alist.Alist)
	aList.UnmarshalJSON([]byte(a))
	aList.Uuid = alist_uuid
	aList.User = user
	err := dal.SaveAlist(*aList)

	aList.Info.Labels = []string{"test1", "test2"}
	err = dal.SaveAlist(*aList)
	// Test breaking
	// Check empty alist.uuid
	aList.Uuid = ""
	err = dal.SaveAlist(*aList)
	assert.Equal(t, i18n.InternalServerErrorMissingAlistUuid, err.Error())
	aList.Uuid = alist_uuid
	// Check empty user.uuid
	aList.User.Uuid = ""
	err = dal.SaveAlist(*aList)
	assert.Equal(t, i18n.InternalServerErrorMissingUserUuid, err.Error())
}

func TestRemoveLabelsForAlistEmptyUuid(t *testing.T) {
	resetDatabase()
	err := dal.RemoveLabelsForAlist("")
	assert.Equal(t, nil, err)
}

func TestGetAndRemoveAlist(t *testing.T) {
	resetDatabase()
	setup := `
INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["english"]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
INSERT INTO user_labels VALUES('english','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO alist_labels VALUES('ada41576-b710-593a-9603-946aaadcb22d','7540fe5f-9847-5473-bdbd-2b20050da0c6','english');
`
	dal.Db.MustExec(setup)

	alist_uuid := "ada41576-b710-593a-9603-946aaadcb22d"
	user_uuid := "7540fe5f-9847-5473-bdbd-2b20050da0c6"

	aList, _ := dal.GetAlist(alist_uuid)
	assert.Equal(t, alist.SimpleList, aList.Info.ListType)

	// Check removing a list of a different user.
	err := dal.RemoveAlist(alist_uuid, "fake")
	assert.Equal(t, i18n.InputDeleteAlistOperationOwnerOnly, err.Error())

	// Check removing a list owned by the user
	err = dal.RemoveAlist(alist_uuid, user_uuid)
	assert.Nil(t, err)
	_, err = dal.GetAlist(alist_uuid)
	assert.Equal(t, i18n.SuccessAlistNotFound, err.Error())
}

func TestGetListsByUserAndLabels(t *testing.T) {
	resetDatabase()
	setup := `
	INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["english"]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
	INSERT INTO user_labels VALUES('english','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO alist_labels VALUES('ada41576-b710-593a-9603-946aaadcb22d','7540fe5f-9847-5473-bdbd-2b20050da0c6','english');
	INSERT INTO alist_labels VALUES('4e075960-5e97-56df-8e1a-c5fe7ea53a44','7540fe5f-9847-5473-bdbd-2b20050da0c6','water');
	INSERT INTO user_labels VALUES('water','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO alist_kv VALUES('4e075960-5e97-56df-8e1a-c5fe7ea53a44','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["water"]},"uuid":"4e075960-5e97-56df-8e1a-c5fe7ea53a44"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	`
	dal.Db.MustExec(setup)

	user_uuid := "7540fe5f-9847-5473-bdbd-2b20050da0c6"
	labels := "english"

	items := dal.GetListsByUserAndLabels(user_uuid, labels)
	assert.Equal(t, 1, len(items))
	items = dal.GetListsByUserAndLabels(user_uuid, "")
	assert.Equal(t, 0, len(items))
	items = dal.GetListsByUserAndLabels(user_uuid, "englishh")
	assert.Equal(t, 0, len(items))

	items = dal.GetListsByUserAndLabels(user_uuid, "water,english")
	assert.Equal(t, 2, len(items))
}

func TestGetListsByUserUuid(t *testing.T) {
	resetDatabase()
	setup := `
	INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["english"]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
	INSERT INTO user_labels VALUES('english','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO alist_labels VALUES('ada41576-b710-593a-9603-946aaadcb22d','7540fe5f-9847-5473-bdbd-2b20050da0c6','english');
	INSERT INTO alist_labels VALUES('4e075960-5e97-56df-8e1a-c5fe7ea53a44','7540fe5f-9847-5473-bdbd-2b20050da0c6','water');
	INSERT INTO user_labels VALUES('water','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	INSERT INTO alist_kv VALUES('4e075960-5e97-56df-8e1a-c5fe7ea53a44','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["water"]},"uuid":"4e075960-5e97-56df-8e1a-c5fe7ea53a44"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
	`
	dal.Db.MustExec(setup)

	user_uuid := "7540fe5f-9847-5473-bdbd-2b20050da0c6"

	items := dal.GetListsByUser(user_uuid)
	assert.Equal(t, 2, len(items))
}
