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
		// TODO needs adding
		if entry.Action != "in-system" {
			return
		}
		dripfeedUUID := entry.UUID
		s.checkForNext(dripfeedUUID)

	case event.ApiDripfeed:
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
		s.checkForNext(dripfeedUUID)
	case event.ApiSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		srsItem := moment.Data

		switch moment.Kind {
		case spaced_repetition.EventKindNew:
			/*
				var settingsInfo SpacedRepetitionSettingsExtID

				json.Unmarshal([]byte(srsItem.Body), &settingsInfo)

				dripfeedUUID := settingsInfo.Settings.ExtID
				if dripfeedUUID == "" {
					return
				}
			*/
			//lastActive := time.Unix(entry.Timestamp, 0).UTC()
			// We could lookup by srs and remove all on the grounds they will not get added
			srsItem := moment.Data
			userUUID := srsItem.UserUUID
			_ = s.repo.DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID, srsItem.UUID)

		case spaced_repetition.EventKindViewed:
			// This can trigger more than one to be added, due to not keeping track of Decrement
			// Only focus on new
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

			// Check if level is 1
			if settings.Level != spaced_repetition.Level1 {
				return
			}

			s.checkForNext(dripfeedUUID)
		}
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

func (s DripfeedService) checkForNext(dripfeedUUID string) {
	// GetNext
	// This is not going to be as easy as I had hoped to access the settings
	// I wonder why I have hidden them

	nextUp, _ := s.repo.GetNext(dripfeedUUID)
	var entry spaced_repetition.ItemInput
	// If body = spaced_repetition.SpacedRepetitionEntry
	switch nextUp.SrsKind {
	case alist.SimpleList:
		entry = spaced_repetition.V1FromDB(string(nextUp.SrsBody))
	case alist.FromToList:
		entry = spaced_repetition.V2FromDB(string(nextUp.SrsBody))
	}

	entry.Reset(time.Now().UTC())
	// TODO I think we don't need this anymore
	// entry.SetExtID(dripfeedUUID)

	// TODO should I check nextUP.UserUUID = userUUID?
	item := spaced_repetition.SpacedRepetitionEntry{
		UserUUID: nextUp.UserUUID,
		UUID:     entry.UUID(),
		Body:     entry.String(),
		WhenNext: entry.WhenNext(),
		Created:  entry.Created(),
	}

	// TODO how do I make sure the dripfeedUUID is not lost
	// Is it simpler to just pass in the repo?
	// Sleep on it
	// Do I send this to spaced_repetition? YES
	// Via the system we trust?
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.SystemSpacedRepetition,
		Data: spaced_repetition.EventSpacedRepetition{
			Kind: spaced_repetition.EventKindNew,
			Data: item,
		},
	})

	_ = s.repo.DeleteByPosition(nextUp.DripfeedUUID, nextUp.Position)
}
