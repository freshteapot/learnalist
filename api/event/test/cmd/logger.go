package main

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/freshteapot/learnalist-api/api/event"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/rifflock/lfshook"
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
	eventLog := log.New()
	eventLog.Out = ioutil.Discard
	eventLog.Formatter = new(log.JSONFormatter)
	eventLog.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
		log.InfoLevel: "/tmp/learnalist.event.test.log",
	}))
	event.New(eventLog)
}
