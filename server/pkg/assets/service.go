package assets

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func NewService(directory string, repo Repository, logEntry *logrus.Entry) AssetService {
	directory = strings.TrimSuffix(directory, "/")
	return AssetService{
		directory: directory,
		logEntry:  logEntry,
		repo:      repo,
	}
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

	return c.File(path)
}

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

	return c.JSON(http.StatusCreated, HttpUploadResponse{
		Href: fmt.Sprintf("/assets/%s/%s%s", userUUID, assetUUID.String(), extension),
	})
}
