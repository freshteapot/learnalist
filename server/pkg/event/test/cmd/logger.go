package main

import (
	"io/ioutil"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func main() {
	setupLogging()

	User := uuid.NewUser()
	PlayList := uuid.NewPlaylist(&User)

	event.Happened(User, event.Type.UserNew)
	event.Happened(PlayList, event.Type.ListNew)
	event.Happened(PlayList, event.Type.ListSaved)
	event.Happened(PlayList, event.Type.ListDeleted)
	event.Todo(PlayList, event.Type.LinkitParse)
}

func setupLogging() {
	eventLog := logrus.New()
	eventLog.Out = ioutil.Discard
	// eventLog.Formatter = new(log.JSONFormatter)
	eventLog.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel: "/tmp/learnalist.event.test.log",
	}, &logrus.JSONFormatter{}))
	event.New(eventLog)
}
