package dripfeed

import (
	"encoding/json"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
// @event.listen: event.SystemSpacedRepetition
// @event.listen: event.ApiDripfeed
// @event.listen: event.ApiSpacedRepetition
func (s DripfeedService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		s.removeUser(entry)
		return
	case event.SystemSpacedRepetition:
		s.handleSystemSpacedRepetitionEvents(entry)
	case event.ApiDripfeed:
		s.handleDripfeedEvents(entry)
	case event.ApiSpacedRepetition:
		s.handleAPISpacedRepetitionEvents(entry)
	}
}

// removeUser when a user is deleted
func (s DripfeedService) removeUser(entry event.Eventlog) {
	if !event.IsUserDeleteEvent(entry) {
		return
	}

	userUUID := entry.UUID
	_ = s.repo.DeleteByUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

// @event.emit: dripfeed.EventDripfeedFinished
// @event.emit: spaced_repetition.EventKindNew
func (s DripfeedService) checkForNext(dripfeedInfo openapi.SpacedRepetitionOvertimeInfo, now time.Time) {
	nextUp, err := s.repo.GetNext(dripfeedInfo.DripfeedUuid)

	if err != nil {
		if err == utils.ErrNotFound {
			// Send event that dripfeedUUID doesnt exist =
			// I wonder if this really means it has finished?
			event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
				Kind: EventDripfeedFinished,
				Data: dripfeedInfo,
			})
			return
		}
		panic(err)
	}

	var entry spaced_repetition.ItemInput
	switch nextUp.SrsKind {
	case alist.SimpleList:
		entry = spaced_repetition.V1FromDB(string(nextUp.SrsBody))
	case alist.FromToList:
		entry = spaced_repetition.V2FromDB(string(nextUp.SrsBody))
	}

	entry.ResetToStart(now)

	item := spaced_repetition.SpacedRepetitionEntry{
		UserUUID: nextUp.UserUUID,
		UUID:     entry.UUID(),
		Body:     entry.String(),
		WhenNext: entry.WhenNext(),
		Created:  entry.Created(),
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.SystemSpacedRepetition,
		Data: spaced_repetition.EventSpacedRepetition{
			Kind: spaced_repetition.EventKindNew,
			Data: item,
		},
		Action: spaced_repetition.EventKindNew,
	})
	// We handle deletion of new entry via the new action event above
}

// @event.emit: dripfeed.EventDripfeedAdded
// @event.emit: dripfeed.EventDripfeedRemoved
func (s DripfeedService) handleDripfeedEvents(entry event.Eventlog) {
	switch entry.Action {
	case event.ActionCreated:
		b, _ := json.Marshal(entry.Data)
		var moment EventDripfeedInputInfo
		json.Unmarshal(b, &moment)

		userUUID := moment.Info.UserUUID
		alistUUID := moment.Info.AlistUUID
		dripfeedUUID := UUID(userUUID, alistUUID)
		items := make([]string, 0)

		now := time.Now().UTC()
		switch moment.Info.Kind {
		case alist.SimpleList:
			settings := spaced_repetition.DefaultSettingsV1(now)
			settings.ExtID = dripfeedUUID

			var input EventDripfeedInputV1
			json.Unmarshal(b, &input)
			for _, listItem := range input.Data {
				item := spaced_repetition.HTTPRequestInputV1{
					Data: listItem,
				}
				b, _ := json.Marshal(item)
				srsItem := spaced_repetition.V1FromPOST(b, settings)
				items = append(items, srsItem.String())
			}

		case alist.FromToList:
			settings := spaced_repetition.DefaultSettingsV2(now)
			settings.ExtID = dripfeedUUID

			var input EventDripfeedInputV2
			json.Unmarshal(b, &input)
			for _, listItem := range input.Data {
				item := spaced_repetition.HTTPRequestInputV2{}
				item.Data = spaced_repetition.HTTPRequestInputV2Item{
					From: listItem.From,
					To:   listItem.To,
				}
				item.Settings.Show = input.Settings.Show

				b, _ := json.Marshal(item)
				srsItem := spaced_repetition.V2FromPOST(b, settings)
				items = append(items, srsItem.String())
			}
		}

		err := s.repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
		if err != nil {
			panic(err)
		}

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: EventDripfeedAdded,
			Data: openapi.SpacedRepetitionOvertimeInfo{
				DripfeedUuid: dripfeedUUID,
				UserUuid:     userUUID,
				AlistUuid:    alistUUID,
			},
		})

		// app_settings.AppendAndSaveSpacedRepetition(storage, userUUID, "123-456")
		// GetNext
		eventTime := time.Unix(entry.Timestamp, 0).UTC()

		info := openapi.SpacedRepetitionOvertimeInfo{
			DripfeedUuid: dripfeedUUID,
			UserUuid:     userUUID,
			AlistUuid:    alistUUID,
		}
		s.checkForNext(info, eventTime)
	case event.ActionDeleted:
		b, _ := json.Marshal(entry.Data)
		var moment EventDripfeedDelete
		json.Unmarshal(b, &moment)
		info, _ := s.repo.GetInfo(moment.DripfeedUUID)
		_ = s.repo.DeleteByUUIDAndUserUUID(info.DripfeedUuid, info.UserUuid)
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: EventDripfeedRemoved,
			Data: info,
		})
	}
}

// @event.listen: spaced_repetition.EventKindNew
// @event.listen: spaced_repetition.EventKindViewed
func (s DripfeedService) handleAPISpacedRepetitionEvents(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var moment spaced_repetition.EventSpacedRepetition
	json.Unmarshal(b, &moment)

	srsItem := moment.Data

	switch moment.Kind {
	case spaced_repetition.EventKindNew:
		srsItem := moment.Data
		userUUID := srsItem.UserUUID
		_ = s.repo.DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID, srsItem.UUID)

	case spaced_repetition.EventKindViewed:
		// This can trigger more than one to be added, due to not keeping track of Decrement
		if moment.Action != spaced_repetition.ActionIncrement {
			return
		}

		var temp SpacedRepetitionSettingsBase

		json.Unmarshal([]byte(srsItem.Body), &temp)
		settings := temp.Settings
		dripfeedUUID := settings.ExtID
		if dripfeedUUID == "" {
			return
		}

		if settings.Level != spaced_repetition.Level1 {
			return
		}

		eventTime := time.Unix(entry.Timestamp, 0).UTC()
		info, _ := s.repo.GetInfo(dripfeedUUID)
		s.checkForNext(info, eventTime)
	}
}

func (s DripfeedService) handleSystemSpacedRepetitionEvents(entry event.Eventlog) {
	if entry.Action != spaced_repetition.EventKindAlreadyInSystem {
		return
	}

	b, _ := json.Marshal(entry.Data)
	var moment spaced_repetition.EventSpacedRepetition
	json.Unmarshal(b, &moment)

	srsItem := moment.Data
	// First we remove the entry from the system that already exists
	err := s.repo.DeleteAllByUserUUIDAndSpacedRepetitionUUID(srsItem.UserUUID, srsItem.UUID)
	if err != nil {
		panic(err)
	}

	var settingsInfo SpacedRepetitionSettingsExtID
	json.Unmarshal([]byte(srsItem.Body), &settingsInfo)
	dripfeedUUID := settingsInfo.Settings.ExtID

	if dripfeedUUID == "" {
		return
	}

	// TODO check if more in system, if not fire event that dripfeed is finished

	eventTime := time.Unix(entry.Timestamp, 0).UTC()
	info, _ := s.repo.GetInfo(dripfeedUUID)
	s.checkForNext(info, eventTime)
}
