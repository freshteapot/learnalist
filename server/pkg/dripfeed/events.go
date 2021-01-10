package dripfeed

import (
	"encoding/json"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/sirupsen/logrus"
)

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

func (s DripfeedService) checkForNext(dripfeedUUID string, now time.Time) {
	nextUp, _ := s.repo.GetNext(dripfeedUUID)
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

func (s DripfeedService) handleDripfeedEvents(entry event.Eventlog) {
	switch entry.Action {
	case event.ActionCreated:
		b, _ := json.Marshal(entry.Data)
		var moment EventDripfeedInputInfo
		json.Unmarshal(b, &moment)

		userUUID := moment.Info.UserUUID
		alistUUID := moment.Info.AlistUUID
		dripfeedUUID := UUID(userUUID, alistUUID)
		items := make([]interface{}, 0)

		now := time.Now().UTC()
		whenNext := now.Add(spaced_repetition.Threshold0)
		settings := spaced_repetition.HTTPRequestInputSettings{
			Level:    spaced_repetition.Level0,
			Created:  now.Format(time.RFC3339),
			WhenNext: whenNext.Format(time.RFC3339),
			ExtID:    dripfeedUUID,
		}

		switch moment.Info.Kind {
		case alist.SimpleList:
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
			// TODO support V2
			panic("TODO")
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
				srsItem := spaced_repetition.V1FromPOST(b, settings)
				items = append(items, srsItem.String())
			}
		}

		err := s.repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
		if err != nil {
			panic(err)
		}

		// GetNext
		eventTime := time.Unix(entry.Timestamp, 0).UTC()
		s.checkForNext(dripfeedUUID, eventTime)
	case event.ActionDeleted:
		b, _ := json.Marshal(entry.Data)
		var moment EventDripfeedDelete
		json.Unmarshal(b, &moment)
		_ = s.repo.DeleteByUUIDAndUserUUID(moment.DripfeedUUID, moment.UserUUID)
	}
}

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
		s.checkForNext(dripfeedUUID, eventTime)
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

	var settingsInfo SpacedRepetitionSettingsExtID
	json.Unmarshal([]byte(srsItem.Body), &settingsInfo)
	dripfeedUUID := settingsInfo.Settings.ExtID

	if dripfeedUUID == "" {
		return
	}

	eventTime := time.Unix(entry.Timestamp, 0).UTC()
	s.checkForNext(dripfeedUUID, eventTime)
}
