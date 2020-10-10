package challenge

type EventChallengeDoneEntry struct {
	Kind     string      `json:"kind"`
	UUID     string      `json:"uuid"`
	UserUUID string      `json:"user_uuid"`
	Data     interface{} `json:"data"`
}

// Event specific
var (
	EventChallengeDone        = "challenge.done"
	EventKindPlank            = "plank"
	EventKindSpacedRepetition = "srs"
)

type EventEntry struct {
	Kind string                  `json:"kind"`
	Data EventChallengeDoneEntry `json:"data"`
}
