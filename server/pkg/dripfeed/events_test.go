package dripfeed_test

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Events", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
	)

	BeforeEach(func() {

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)

		logger, _ = test.NewNullLogger()
		fmt.Println(logger)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "dripfeedService", mock.Anything)
	})

	It("REMOVE", func() {
		Expect("1").To(Equal("1"))
	})

	It("Making sure things work", func() {
		type eventDripfeedInput struct {
			UserUUID string      `json:"user_uuid"`
			Kind     string      `json:"kind"` // This is the list_type, at some point I will drop list_type :P
			Data     interface{} `json:"data"` // TODO I think with openapi-generator we might be able to move to something else.
		}
		raw := `{"kind":"api.dripfeed","data":{"user_uuid":"7197b389-cfe6-4fa8-9aea-98d49b305039","kind":"v1","data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"]},"timestamp":1610273658,"action":"created"}`

		var entry event.Eventlog
		json.Unmarshal([]byte(raw), &entry)
		b, _ := json.Marshal(entry.Data)

		var moment dripfeed.EventDripfeedInputV1
		json.Unmarshal(b, &moment)
		fmt.Println(moment.Data)
	})

	FIt("Check if new has dripfeed", func() {
		raw := `{"kind":"api.spacedrepetition","data":{"kind":"new","data":{"uuid":"bfe3cc8ad82c1e8282b53df0a7a78685042d9f5b","body":"{\"show\":\"monday\",\"kind\":\"v1\",\"uuid\":\"bfe3cc8ad82c1e8282b53df0a7a78685042d9f5b\",\"data\":\"monday\",\"settings\":{\"level\":\"0\",\"when_next\":\"2021-01-10T15:37:28Z\",\"created\":\"2021-01-10T14:37:28Z\",\"ext_id\":\"b17ef2deb2d1836dfe534de67e710e23c5b67e88\"}}","user_uuid":"4eccc98d-90ea-42ba-84d4-d0688b64d24e","when_next":"2021-01-10T15:37:28Z","created":"2021-01-10T14:37:28Z"}},"timestamp":1610289448}`
		var entry event.Eventlog
		json.Unmarshal([]byte(raw), &entry)

		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		srsItem := moment.Data

		var info dripfeed.SpacedRepetitionSettingsExtID

		json.Unmarshal([]byte(srsItem.Body), &info)
		fmt.Println(info.Settings.ExtID)

	})
})
