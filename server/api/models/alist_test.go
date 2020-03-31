package models

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
)

func (suite *ModelSuite) TestSaveAlistPost() {
	userUUID := suite.UserUUID
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

	var aList alist.Alist

	aList.UnmarshalJSON([]byte(a))
	aList.User.Uuid = userUUID

	// TODO move this into own test
	aList, err := dal.SaveAlist(http.MethodPost, aList)
	suite.NoError(err)

	aList, err = dal.SaveAlist(http.MethodPut, aList)
	suite.NoError(err)

	aList.Info.Labels = []string{"test1", "test2"}
	aList, err = dal.SaveAlist(http.MethodPut, aList)
	suite.NoError(err)
	suite.Equal(2, len(aList.Info.Labels))
}

func (suite *ModelSuite) TestSaveAListInternalIssues() {
	var err error
	userUUID := suite.UserUUID
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

	var aList alist.Alist
	aList.UnmarshalJSON([]byte(a))

	// Check empty user.uuid
	aList.User.Uuid = ""
	_, err = dal.SaveAlist(http.MethodPost, aList)
	suite.Equal(i18n.InternalServerErrorMissingUserUuid, err.Error())

	_, err = dal.SaveAlist(http.MethodPut, aList)
	suite.Equal(i18n.InternalServerErrorMissingUserUuid, err.Error())
	// Check empty alist Uuid
	aList.User.Uuid = userUUID
	aList.Uuid = ""
	_, err = dal.SaveAlist(http.MethodPut, aList)
	suite.Equal(i18n.InternalServerErrorMissingAlistUuid, err.Error())
}

func (suite *ModelSuite) TestSaveAListViaPutWithSameData() {
	userUUID := suite.UserUUID
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
	var aList alist.Alist
	aList.UnmarshalJSON([]byte(a))
	aList.User.Uuid = userUUID
	aList, err := dal.SaveAlist(http.MethodPost, aList)
	suite.NoError(err)
	aUUID := aList.Uuid
	aList, err = dal.SaveAlist(http.MethodPut, aList)
	suite.NoError(err)
	bUUID := aList.Uuid
	aList, err = dal.SaveAlist(http.MethodPut, aList)
	suite.NoError(err)
	cUUID := aList.Uuid
	// Make sure the uuid is the same
	suite.Equal(aUUID, bUUID)
	suite.Equal(bUUID, cUUID)
}

func (suite *ModelSuite) TestSaveAListViaPutWithNotFoundUuid() {
	userUUID := suite.UserUUID
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
		    },
				"uuid": "fake"
		}
		`

	var aList alist.Alist
	aList.UnmarshalJSON([]byte(a))
	aList.User.Uuid = userUUID
	aList, err := dal.SaveAlist(http.MethodPut, aList)
	suite.Equal(i18n.SuccessAlistNotFound, err.Error())
}

func (suite *ModelSuite) TestSaveAListEmptyList() {
	userUUID := suite.UserUUID
	var input alist.Alist
	input.User.Uuid = userUUID

	aList, err := dal.SaveAlist(http.MethodPost, input)
	suite.Equal(alist.Alist{}, aList)
	suite.Equal(fmt.Sprintf(i18n.ValidationErrorList, "Title cannot be empty.\nInvalid option for info.shared_with"), err.Error())
}

func (suite *ModelSuite) TestRemoveLabelsForAlistEmptyUuid() {
	err := dal.RemoveLabelsForAlist("")
	suite.Equal(nil, err)
}

func (suite *ModelSuite) TestGetAndRemoveAlist() {
	userUUID := suite.UserUUID
	setup := `
INSERT INTO alist_kv VALUES('ada41576-b710-593a-9603-946aaadcb22d','v1','{"data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"],"info":{"title":"Days of the Week","type":"v1","labels":["english"]},"uuid":"ada41576-b710-593a-9603-946aaadcb22d"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO user_labels VALUES('english','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO alist_labels VALUES('ada41576-b710-593a-9603-946aaadcb22d','7540fe5f-9847-5473-bdbd-2b20050da0c6','english');
`
	dal.Db.MustExec(setup)

	alistUUID := "ada41576-b710-593a-9603-946aaadcb22d"

	aList, _ := dal.GetAlist(alistUUID)
	suite.Equal(alist.SimpleList, aList.Info.ListType)

	// Check removing a list of a different user.
	err := dal.RemoveAlist(alistUUID, "fake")
	suite.Equal(i18n.InputDeleteAlistOperationOwnerOnly, err.Error())

	// Check removing a list owned by the user
	err = dal.RemoveAlist(alistUUID, userUUID)
	suite.Nil(err)
	_, err = dal.GetAlist(alistUUID)
	suite.Equal(i18n.SuccessAlistNotFound, err.Error())
}

func (suite *ModelSuite) TestGetListsByUserWithFilters() {
	userUUID := suite.UserUUID
	setup := `
INSERT INTO alist_kv VALUES('0cf0f9de-c18f-52d5-8352-cee1ab7eab28','v1','{"data":["car"],"info":{"title":"Days of the Week","type":"v1","labels":[]},"uuid":"0cf0f9de-c18f-52d5-8352-cee1ab7eab28"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO alist_kv VALUES('45bc50f7-1228-5bc2-9daa-2be6b6fbd1de','v2','{"data":[{"from":"car","to":"bil"}],"info":{"title":"Days of the Week","type":"v2","labels":[]},"uuid":"45bc50f7-1228-5bc2-9daa-2be6b6fbd1de"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO alist_kv VALUES('292a4fd3-8835-5435-9e68-7085ab901730','v1','{"data":["car"],"info":{"title":"Days of the Week","type":"v1","labels":["car"]},"uuid":"292a4fd3-8835-5435-9e68-7085ab901730"}','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO user_labels VALUES('car','7540fe5f-9847-5473-bdbd-2b20050da0c6');
INSERT INTO alist_labels VALUES('292a4fd3-8835-5435-9e68-7085ab901730','7540fe5f-9847-5473-bdbd-2b20050da0c6','car');
`
	dal.Db.MustExec(setup)
	items := dal.GetListsByUserWithFilters(userUUID, "", "")
	suite.Equal(3, len(items))

	items = dal.GetListsByUserWithFilters(userUUID, "car", "")
	suite.Equal(1, len(items))
	suite.Equal("292a4fd3-8835-5435-9e68-7085ab901730", items[0].Uuid)

	items = dal.GetListsByUserWithFilters(userUUID, "", "v1")
	suite.Equal(2, len(items))

	items = dal.GetListsByUserWithFilters(userUUID, "", "v2")
	suite.Equal(1, len(items))
	suite.Equal("45bc50f7-1228-5bc2-9daa-2be6b6fbd1de", items[0].Uuid)

	items = dal.GetListsByUserWithFilters(userUUID, "car", "v1")
	suite.Equal(1, len(items))
	suite.Equal("292a4fd3-8835-5435-9e68-7085ab901730", items[0].Uuid)
}
