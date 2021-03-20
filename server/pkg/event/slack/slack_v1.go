package slack

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type PostWebhook = func(url string, msg *slack.WebhookMessage) error

type SlackEvents struct {
	post       PostWebhook
	webhook    string
	logContext logrus.FieldLogger
}

func NewSlackV1Events(post PostWebhook, webhook string, logContext logrus.FieldLogger) SlackEvents {
	return SlackEvents{
		post:       post,
		webhook:    webhook,
		logContext: logContext,
	}
}

// @event.listen: event.ApiUserRegister
// @event.listen: event.ApiUserLogin
// @event.listen: event.ApiUserLogout
// @event.listen: event.BrowserUserLogout
// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
// @event.listen: event.ApiListSaved
// @event.listen: event.ApiListDelete
// @event.listen: event.ApiSpacedRepetition
// @event.listen: event.ApiPlank
// @event.listen: challenge.EventChallengeDone
// @event.listen: challenge.EventChallengeNewRecord
func (s SlackEvents) Read(entry event.Eventlog) {
	var msg slack.WebhookMessage

	switch entry.Kind {
	case event.ApiUserRegister:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventNewUser
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("%s: user:%s registered via %s", entry.Kind, moment.UUID, moment.Kind)
	case event.ApiUserLogin:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventUser
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("%s: user:%s logged in via %s", entry.Kind, moment.UUID, moment.Kind)
	case event.ApiUserLogout:
		fallthrough
	case event.BrowserUserLogout:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventUser
		json.Unmarshal(b, &moment)
		via := "api"
		if entry.Kind == event.BrowserUserLogout {
			via = "browser"
		}

		clearing := "current session"
		if moment.Kind == event.KindUserLogoutSessions {
			clearing = "all sessions"
		}

		msg.Text = fmt.Sprintf("%s: user:%s logged out via %s, clearing %s", entry.Kind, moment.UUID, via, clearing)
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		userUUID := entry.UUID
		msg.Text = fmt.Sprintf("%s: user:%s should be deleted", entry.Kind, userUUID)
	case event.ApiListSaved:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf(`list:%s (%s) %s by user:%s`, moment.UUID, moment.Data.Info.SharedWith, moment.Action, moment.UserUUID)
	case event.ApiListDelete:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)
		// TODO https://github.com/freshteapot/learnalist-api/issues/212
		msg.Text = fmt.Sprintf("list:%s deleted by user:%s", moment.UUID, moment.UserUUID)
	case event.ApiSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		if moment.Kind == spaced_repetition.EventKindNew {
			msg.Text = fmt.Sprintf("user:%s added a new entry for spaced based learning", moment.Data.UserUUID)
		}

		if moment.Kind == spaced_repetition.EventKindViewed {
			when := "na"
			if moment.Action == "incr" {
				when = "later"
			}

			if moment.Action == "decr" {
				when = "sooner"
			}
			msg.Text = fmt.Sprintf("user:%s will be reminded %s of entry:%s", moment.Data.UserUUID, when, moment.Data.UUID)
		}

		if moment.Kind == spaced_repetition.EventKindDeleted {
			msg.Text = fmt.Sprintf("user:%s removed entry:%s from spaced based learning", moment.Data.UserUUID, moment.Data.UUID)
		}
	case event.ApiSpacedRepetitionOvertime:
		switch entry.Action {
		case event.ActionCreated:
			b, _ := json.Marshal(entry.Data)
			var data dripfeed.EventDripfeedInputInfo
			json.Unmarshal(b, &data)
			msg.Text = fmt.Sprintf("spaced repetition over time created uuid:%s, user:%s, list:%s", entry.UUID, data.Info.UserUUID, data.Info.AlistUUID)

		case event.ActionDeleted:
			b, _ := json.Marshal(entry.Data)
			var data openapi.SpacedRepetitionOvertimeInfo
			json.Unmarshal(b, &data)
			msg.Text = fmt.Sprintf("spaced repetition over time deleted uuid:%s, user:%s, list:%s", data.DripfeedUuid, data.UserUuid, data.AlistUuid)
		default:
			msg.Text = fmt.Sprintf(`%s action not supported for %s`, entry.Action, entry.Kind)
		}

	case event.SystemSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		switch moment.Kind {
		case spaced_repetition.EventKindNew:
			msg.Text = fmt.Sprintf("spaced repetition over time system added entry:%s for user:%s", entry.UUID, moment.Data.UserUUID)
		case spaced_repetition.EventKindAlreadyInSystem:
			msg.Text = fmt.Sprintf("spaced repetition over time system added entry:%s for user:%s that already exists", entry.UUID, moment.Data.UserUUID)
		default:
			msg.Text = fmt.Sprintf(`%s kind not supported for %s`, moment.Kind, entry.Kind)
		}

	case dripfeed.EventDripfeedAdded:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("spaced repetition over time activated for user:%s from list:%s", moment.UserUuid, moment.AlistUuid)
	case dripfeed.EventDripfeedRemoved:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("spaced repetition over time stopped for user:%s from list:%s", moment.UserUuid, moment.AlistUuid)
	case dripfeed.EventDripfeedFinished:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("spaced repetition over time finished for user:%s from list:%s", moment.UserUuid, moment.AlistUuid)
	case event.ApiPlank:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventPlank
		json.Unmarshal(b, &moment)

		switch moment.Action {
		case event.ActionNew:
			msg.Text = fmt.Sprintf("user:%s added a plank:%s", moment.UserUUID, moment.Data.Uuid)
		case event.ActionDeleted:
			msg.Text = fmt.Sprintf("user:%s deleted plank:%s", moment.UserUUID, moment.Data.Uuid)
		default:
			msg.Text = fmt.Sprintf(`%s action not supported %s`, moment.Action, entry.Kind)
		}

	case challenge.EventChallengeDone:
		b, _ := json.Marshal(entry.Data)
		var moment challenge.EventChallengeDoneEntry
		json.Unmarshal(b, &moment)
		if moment.Kind == challenge.EventKindPlank {
			b, _ = json.Marshal(moment.Data)
			var record openapi.Plank
			json.Unmarshal(b, &record)
			msg.Text = fmt.Sprintf("user:%s added a plank:%s to challenge:%s", moment.UserUUID, record.Uuid, moment.UUID)
		} else {
			return
		}
		// TODO Add challenge notification
	case challenge.EventChallengeNewRecord:
		msg.Text = s.challengeNewRecord(entry)
		if msg.Text == "" {
			return
		}
	case challenge.EventChallengeCreated:
		var momentKV event.EventKV
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &momentKV)
		b, _ = json.Marshal(momentKV.Data)
		var moment challenge.ChallengeInfo
		json.Unmarshal(b, &moment)

		msg.Text = fmt.Sprintf("user:%s created challenge:%s", moment.CreatedBy, moment.UUID)
	//case challenge.EventChallengeDeleted   = "challenge.deleted"
	case challenge.EventChallengeJoined:
		var momentKV event.EventKV
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &momentKV)
		b, _ = json.Marshal(momentKV.Data)
		var moment challenge.ChallengeJoined
		json.Unmarshal(b, &moment)

		msg.Text = fmt.Sprintf("user:%s joined challenge:%s", moment.UserUUID, moment.UUID)
	case challenge.EventChallengeLeft:
		var momentKV event.EventKV
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &momentKV)
		b, _ = json.Marshal(momentKV.Data)
		var moment challenge.ChallengeLeft
		json.Unmarshal(b, &moment)

		msg.Text = fmt.Sprintf("user:%s left challenge:%s", moment.UserUUID, moment.UUID)
	case event.MobileDeviceRegistered:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.MobileDeviceInfo
		json.Unmarshal(b, &moment)

		userUUID := moment.UserUuid
		msg.Text = fmt.Sprintf(`user:%s registered mobile token for app:"%s"`, userUUID, moment.AppIdentifier)
	case event.MobileDeviceRemove:
		msg.Text = "Removing a token based on feedback from fcm, a follow up event should happen."
	case event.MobileDeviceRemoved:
		b, _ := json.Marshal(entry.Data)
		var deviceInfo openapi.MobileDeviceInfo
		json.Unmarshal(b, &deviceInfo)
		msg.Text = fmt.Sprintf(`user:%s fcm token from app:"%s" has been removed`, deviceInfo.UserUuid, deviceInfo.AppIdentifier)
	case remind.EventApiRemindDailySettings:
		b, _ := json.Marshal(entry.Data)
		var settings openapi.RemindDailySettings
		json.Unmarshal(b, &settings)

		userUUID := entry.UUID
		switch entry.Action {
		case event.ActionDeleted:
			msg.Text = fmt.Sprintf(`user:%s removed daily reminder for app:"%s"`, userUUID, settings.AppIdentifier)
		case event.ActionUpsert:
			msg.Text = fmt.Sprintf(`user:%s setup daily reminder for app:"%s"`, userUUID, settings.AppIdentifier)
		default:
			msg.Text = fmt.Sprintf(`%s action not supported %s`, entry.Kind, entry.Action)
		}
	case event.ApiAppSettingsRemindV1:
		//b, _ := json.Marshal(entry.Data)
		//var settings openapi.AppSettingsRemindV1
		//json.Unmarshal(b, &settings)

		userUUID := entry.UUID
		switch entry.Action {
		case event.ActionUpsert:
			msg.Text = fmt.Sprintf(`user:%s updated app settings for app:remind_v1`, userUUID)
		default:
			msg.Text = fmt.Sprintf(`%s action not supported %s`, entry.Kind, entry.Action)
		}
	default:
		msg.Text = entry.Kind
	}

	// We parse this in to make it easier to mock
	err := s.post(s.webhook, &msg)
	if err != nil {
		s.logContext.Panic(err)
	}
}

func (s SlackEvents) challengeNewRecord(entry event.Eventlog) string {
	// TODO move / copy to the system that sends push notifications
	// TODO use this to trigger a rebuild of the challenge page for static site
	// Use this event to add user to active list
	if entry.Kind != challenge.EventChallengeNewRecord {
		return ""
	}

	var moment challenge.EventChallengeDoneEntry
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &moment)

	challengeUUID := moment.UUID

	b, _ = json.Marshal(moment.Data)
	var record challenge.ChallengeRecordUUID
	json.Unmarshal(b, &record)

	return fmt.Sprintf("user:%s added record:%s (%s) to challenge:%s\n",
		moment.UserUUID,
		record.UUID,
		moment.Kind,
		challengeUUID,
	)
}
