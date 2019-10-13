package api_test

//2, as I am being really lazy :(, once all moved over to ginkgo remove.
import (
	"net/http"

	mockHugo "github.com/freshteapot/learnalist-api/server/alists/pkg/hugo/mocks"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	mockModels "github.com/freshteapot/learnalist-api/server/api/models/mocks"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Alist endpoints", func() {
	AfterEach(emptyDatabase)

	When("Basic crud", func() {
		var userUUID string
		BeforeEach(func() {
			testHugoHelper := new(mockHugo.HugoSiteBuilder)
			testHugoHelper.On("Write", mock.Anything)
			testHugoHelper.On("Remove", mock.Anything)
			m.HugoHelper = testHugoHelper

			//inputUserA := getValidUserRegisterInput("a")
			//userUUID, _ = createNewUserWithSuccess(inputUserA)
		})
		AfterEach(func() {

		})
		Context("Save a list", func() {
			var datastore *mockModels.Datastore
			BeforeEach(func() {
				datastore = &mockModels.Datastore{}
				m.Datastore = datastore
			})
			It("Get, is not accepted", func() {
				user := &uuid.User{
					Uuid: userUUID,
				}
				//datastore.On("GetUserByCredentials", mock.Anything).Return(user, nil)

				input := ""
				req, rec := setupFakeEndpoint(http.MethodGet, "/api/v1/alist", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
			})

			It("Post, success", func() {
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(alist.NewTypeV1(), nil)
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}
				//datastore.On("GetUserByCredentials", mock.Anything).Return(user, nil)

				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/alist", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusCreated))
			})
		})
	})

})
