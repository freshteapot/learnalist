package hugo

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/sirupsen/logrus"
)

func (h *HugoHelper) WriteList(aList alist.Alist) {
	h.AlistWriter.Data(aList)
	h.AlistWriter.Content(aList)
}

// WriteListsByUser
func (h *HugoHelper) WriteListsByUser(userUUID string, lists []alist.ShortInfo) {
	h.AlistsByUserWriter.Data(userUUID, lists)
	h.AlistsByUserWriter.Content(userUUID)
}

func (h *HugoHelper) WritePublicLists(lists []alist.ShortInfo) {
	h.PublicListsWriter.Data(lists)
}

type hugoFileWriter struct {
	logContext logrus.FieldLogger
}

func NewHugoFileWriter(logContext logrus.FieldLogger) FileWriter {
	return hugoFileWriter{
		logContext: logContext,
	}
}

func (w hugoFileWriter) Write(path string, data []byte) {
	logContext := w.logContext.WithFields(logrus.Fields{
		"event": "write-file",
	})
	err := ioutil.WriteFile(path, data, 0744)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"path":  path,
			"error": err,
		}).Error("Writing file")
		return
	}

	//logContext.WithFields(logrus.Fields{
	//	"path": path,
	//}).Info("file written")
}

func (w hugoFileWriter) Remove(path string) {
	logContext := w.logContext.WithFields(logrus.Fields{
		"event": "delete-file",
	})

	// Command to remove lists in hugo that are no longer in the DB
	// https://github.com/freshteapot/learnalist-api/issues/98

	err := os.Remove(path)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "no such file or directory") {
			logContext.WithFields(logrus.Fields{
				"path":  path,
				"error": err,
			}).Error("file removed")
			return
		}
	}

	//logContext.WithFields(logrus.Fields{
	//	"path": path,
	//}).Info("file removed")
}
