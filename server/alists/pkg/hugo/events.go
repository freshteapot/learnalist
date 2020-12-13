package hugo

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
)

func (h HugoHelper) ListenForEvents() {
	// TODO this will mean we move to nats, or at least away from memory
	// TODO setup config for the stream, short duration, bigger payload
	event.GetBus().Subscribe(event.TopicStaticSite, "hugoHelper", h.OnEvent)
}

func (h HugoHelper) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ChangesetChallenge:
		b, _ := json.Marshal(entry.Data)
		var moment challenge.ChallengeInfo
		json.Unmarshal(b, &moment)

		if moment.UUID == "" {
			// Should never happen
			return
		}

		if entry.Action == event.ActionDeleted {
			challengeUUID := moment.UUID
			h.challengeWriter.Remove(challengeUUID)
			h.RegisterCronJob()
			return
		}

		// Write challenge
		h.challengeWriter.Data(moment)
		h.challengeWriter.Content(moment)
		h.RegisterCronJob()
		return
	default:
		return
	}
}
