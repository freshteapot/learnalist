package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func NewService(directory string, acl acl.AclAsset, repo Repository, logEntry *logrus.Entry) AssetService {
	directory = strings.TrimSuffix(directory, "/")
	s := AssetService{
		acl:       acl,
		directory: directory,
		logEntry:  logEntry,
		repo:      repo,
	}

	//event.GetBus().Subscribe(event.TopicMonolog, func(entry event.Eventlog) {
	event.GetBus().Listen(func(entry event.Eventlog) {
		if entry.Kind != event.ApiUserDelete {
			return
		}

		b, err := json.Marshal(entry.Data)
		if err != nil {
			return
		}

		var moment event.EventUser
		json.Unmarshal(b, &moment)
		s.DeleteUserAssets(moment.UUID)
	})
	return s
}

func (s *AssetService) DeleteUserAssets(userUUID string) {
	if userUUID == "" {
		return
	}

	directory := fmt.Sprintf("%s/%s", s.directory, userUUID)
	s.logEntry.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"directory": directory,
		"action":    "rm_user_assets",
	}).Info("removing assets")
	os.RemoveAll(directory)
}

func (s *AssetService) InitCheck() {
	logEntry := s.logEntry.WithField("event", "init-check")

	fakeUserUUID := "fake-user"
	directory := fmt.Sprintf("%s/%s", s.directory, fakeUserUUID)
	err := os.MkdirAll(directory, 0744)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "create_directory",
		}).Fatal("init check")
	}

	tempFile := fmt.Sprintf("%s/temp.txt", directory)
	file, err := os.Create(tempFile)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "write_temp_file",
		}).Fatal("init check")
	}
	defer file.Close()

	err = os.RemoveAll(directory)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "clean_up",
		}).Fatal("init check")
	}

	logEntry.Info("✔ assets service check")
}

// DeleteEntry Deletes a single entry based on the UUID
func (s *AssetService) DeleteEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	UUID := c.Param("uuid")

	if UUID == "" {
		response := api.HttpResponseMessage{
			Message: "Missing asset uuid",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// Loook up asset
	asset, err := s.repo.GetEntry(UUID)
	if err != nil {
		if err == ErrNotFound {
			response := api.HttpResponseMessage{
				Message: "Asset not found",
			}
			return c.JSON(http.StatusNotFound, response)
		}
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if asset.UserUUID != userUUID {
		response := api.HttpResponseMessage{
			Message: "Access denied",
		}
		return c.JSON(http.StatusForbidden, response)
	}

	err = s.acl.DeleteAsset(UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	err = s.repo.DeleteEntry(userUUID, UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *AssetService) Share(c echo.Context) error {
	response := api.HttpResponseMessage{
		Message: "",
	}

	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	var input openapi.HttpAssetShareRequestBody

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response.Message = "Check the documentation"
		return c.JSON(http.StatusBadRequest, response)
	}

	sharedWith := input.Action

	allowed := []string{aclKeys.SharedWithPublic, aclKeys.NotShared}
	if !utils.StringArrayContains(allowed, sharedWith) {
		response.Message = "Check the documentation"
		return c.JSON(http.StatusBadRequest, response)
	}

	// Loook up asset
	asset, _ := s.repo.GetEntry(input.Uuid)
	if asset.UserUUID != userUUID {
		response := api.HttpResponseMessage{
			Message: "Access denied",
		}
		return c.JSON(http.StatusForbidden, response)
	}

	// Update who it is shared with
	switch sharedWith {
	case aclKeys.SharedWithPublic:
		err = s.acl.ShareAssetWithPublic(asset.UUID)
	case aclKeys.NotShared:
		err = s.acl.MakeAssetPrivate(asset.UUID, userUUID)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	response.Message = "Updated"
	return c.JSON(http.StatusOK, response)
}

func (s *AssetService) GetAsset(c echo.Context) error {
	asset := strings.TrimPrefix(c.Request().URL.Path, "/assets/")
	// This might do nothing, due to echo routing
	if strings.Contains(asset, "../") {
		return c.NoContent(http.StatusTeapot)
	}

	path := fmt.Sprintf("%s/%s", s.directory, asset)
	_, err := os.Stat(path)

	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	user := c.Get("loggedInUser")
	userUUID := ""
	if user != nil {
		userUUID = user.(uuid.User).Uuid
	}

	extUUID := ""
	parts := strings.Split(asset, "/")
	if len(parts) == 2 {
		// Ugly code
		extUUID = parts[1]
		parts = strings.Split(extUUID, ".")
		extUUID = parts[0]
	}

	allow, err := s.acl.HasUserAssetReadAccess(extUUID, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if !allow {
		response := api.HttpResponseMessage{
			Message: "Access denied",
		}
		return c.JSON(http.StatusForbidden, response)
	}

	return c.File(path)
}

// TODO add privacy option private, public
// ln -s /tmp/learnalist/assets /Users/tinkerbell/git/learnalist-api/hugo/public/assets
// https://echo.labstack.com/cookbook/file-upload
func (s *AssetService) Upload(c echo.Context) error {
	logEntry := s.logEntry
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	assetUUID := guuid.New()
	directory := fmt.Sprintf("%s/%s", s.directory, userUUID)

	logEntry = logEntry.WithFields(logrus.Fields{
		"asset_uuid": assetUUID.String(),
		"user_uuid":  userUUID,
		"directory":  directory,
	})

	sharedWith := c.FormValue("shared_with")
	if sharedWith == "" {
		sharedWith = aclKeys.NotShared
	}

	allowed := []string{aclKeys.SharedWithPublic, aclKeys.NotShared}
	if !utils.StringArrayContains(allowed, sharedWith) {
		response := api.HttpResponseMessage{
			Message: "Check the documentation",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	err := os.MkdirAll(directory, 0744)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "create_directory",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	file, err := c.FormFile("file")

	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "form_file",
		}).Error("asset upload")
		response := api.HttpResponseMessage{
			Message: "Check the documentation",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	src, err := file.Open()
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "open_file",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	defer src.Close()

	fileHeader := make([]byte, 512)
	if _, err := src.Read(fileHeader); err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "read_mimetype",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	// Reset so we save the whole file
	src.Seek(0, io.SeekStart)

	mimeType := http.DetectContentType(fileHeader)
	mimeTypes, _ := mime.ExtensionsByType(mimeType)
	if len(mimeTypes) == 0 {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "reading_mimetype_from_file",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	extension := mimeTypes[0]

	path := fmt.Sprintf("%s/%s%s", directory, assetUUID.String(), extension)

	logEntry = logEntry.WithField("path", path)
	dst, err := os.Create(path)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "create_asset",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	defer dst.Close()

	// Copy uploaded file to path
	if _, err = io.Copy(dst, src); err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "save_asset_to_disk",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	logEntry.WithFields(logrus.Fields{
		"action": "uploaded",
	}).Info("asset upload")

	// write to db
	err = s.repo.SaveEntry(AssetEntry{
		UUID:      assetUUID.String(),
		UserUUID:  userUUID,
		Extension: extension,
	})

	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":     err,
			"mime_type": extension,
			"action":    "failed_to_save_db",
		}).Error("asset upload")
		// Try to clean up
		os.Remove(path)
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	// Save access
	switch sharedWith {
	case aclKeys.SharedWithPublic:
		err = s.acl.ShareAssetWithPublic(assetUUID.String())
	case aclKeys.NotShared:
		err = s.acl.MakeAssetPrivate(assetUUID.String(), userUUID)
	}

	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"error":  err,
			"action": "failed_to_set_shared_with_to_db",
		}).Error("asset upload")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.JSON(http.StatusCreated, openapi.HttpAssetUploadResponse{
		Href: fmt.Sprintf("/assets/%s/%s%s", userUUID, assetUUID.String(), extension),
		Uuid: assetUUID.String(),
		Ext:  extension,
	})
}
