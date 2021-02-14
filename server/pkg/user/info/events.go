package info

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
)

// @event.listen: dripfeed.EventDripfeedAdded
// @event.listen: dripfeed.EventDripfeedRemoved
// @event.listen: dripfeed.EventDripfeedFinished
func (s UserInfoService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case dripfeed.EventDripfeedAdded:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		AppendAndSaveSpacedRepetition(s.userRepo, moment.UserUuid, moment.AlistUuid)
	case dripfeed.EventDripfeedRemoved:
		fallthrough
	case dripfeed.EventDripfeedFinished:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		RemoveAndSaveSpacedRepetition(s.userRepo, moment.UserUuid, moment.AlistUuid)
	}
}
