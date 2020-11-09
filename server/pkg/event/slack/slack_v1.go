package event

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type SlackEvents struct {
	webhook    string
	logContext logrus.FieldLogger
}

func NewSlackV1Events(webhook string, logContext logrus.FieldLogger) SlackEvents {
	return SlackEvents{
		webhook:    webhook,
		logContext: logContext,
	}
}

func (s SlackEvents) Read(entry event.Eventlog) {
	var msg slack.WebhookMessage

	switch entry.Kind {
	case event.ApiUserRegister:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventUser
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("%s: user %s registered via %s", entry.Kind, moment.UUID, moment.Kind)
	case event.ApiUserLogin:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventUser
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("%s: user %s logged in via %s", entry.Kind, moment.UUID, moment.Kind)
	case event.ApiUserLogout:
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

		msg.Text = fmt.Sprintf("%s: user %s logged out via %s, clearing %s", entry.Kind, moment.UUID, via, clearing)
	case event.ApiUserDelete:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventUser
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("%s: user %s should be deleted", entry.Kind, moment.UUID)
	case event.ApiListSaved:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf(`list:%s (%s) %s by user:%s`, moment.UUID, moment.Data.Info.SharedWith, moment.Action, moment.UserUUID)
	case event.ApiListDelete:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)
		msg.Text = fmt.Sprintf("list:%s deleted by user:%s", moment.UUID, moment.UserUUID)
	case spaced_repetition.EventApiSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		if moment.Kind == spaced_repetition.EventKindNew {
			msg.Text = fmt.Sprintf("User:%s added a new entry for spaced based learning", moment.Data.UserUUID)
		}

		if moment.Kind == spaced_repetition.EventKindViewed {
			when := "na"
			if moment.Action == "incr" {
				when = "later"
			}

			if moment.Action == "decr" {
				when = "sooner"
			}
			msg.Text = fmt.Sprintf("User:%s will be reminded %s of entry:%s", moment.Data.UserUUID, when, moment.Data.UUID)
		}

		if moment.Kind == spaced_repetition.EventKindDeleted {
			msg.Text = fmt.Sprintf("User:%s removed entry:%s from spaced based learning", moment.Data.UserUUID, moment.Data.UUID)
		}
	case plank.EventApiPlank:
		b, _ := json.Marshal(entry.Data)
		var moment plank.EventPlank
		json.Unmarshal(b, &moment)
		if moment.Kind == plank.EventKindNew {
			msg.Text = fmt.Sprintf("User:%s added a plank:%s", moment.UserUUID, moment.Data.UUID)
		}

		if moment.Kind == plank.EventKindDeleted {
			msg.Text = fmt.Sprintf("User:%s deleted a plank:%s", moment.UserUUID, moment.Data.UUID)
		}
	case challenge.EventChallengeDone:
		b, _ := json.Marshal(entry.Data)
		var moment challenge.EventChallengeDoneEntry
		json.Unmarshal(b, &moment)
		if moment.Kind == challenge.EventKindPlank {
			b, _ = json.Marshal(moment.Data)
			var record plank.HttpRequestInput
			json.Unmarshal(b, &record)
			msg.Text = fmt.Sprintf("User:%s added a plank:%s to challenge:%s", moment.UserUUID, record.UUID, moment.UUID)
		} else {
			return
		}
		// TODO Add challenge notification
	case challenge.EventChallengeNewRecord:
		msg.Text = s.challengeNewRecord(entry)
		if msg.Text == "" {
			return
		}

	default:
		b, _ := json.Marshal(entry)
		fmt.Println(string(b))
		msg.Text = entry.Kind
	}

	err := slack.PostWebhook(s.webhook, &msg)
	if err != nil {
		s.logContext.Panic(err)
	}
}

func (s SlackEvents) challengeNewRecord(entry event.Eventlog) string {
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

	return fmt.Sprintf("Challenge %s (%s) has a new record %s by user %s\n",
		challengeUUID,
		moment.Kind,
		record.UUID,
		moment.UserUUID)
}
