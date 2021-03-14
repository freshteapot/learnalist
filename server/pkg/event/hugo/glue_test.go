package hugo_test

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Hugo Glue", func() {
	It("", func() {
		var entry event.Eventlog
		json.Unmarshal([]byte(`{"kind":"api.list.delete","data":{"uuid":"a030e570-ab6b-524d-815d-8aa23968a74e","user_uuid":"0098bc8b-42c0-469e-83e3-8f3892a42c37","data":{"data":null,"info":{"title":"","type":"","labels":null,"shared_with":""},"uuid":""}},"timestamp":1615729458}`), &entry)

		b, _ := json.Marshal(entry.Data)
		var moment event.EventListOwner
		err := json.Unmarshal(b, &moment)
		fmt.Println("err", err)
		fmt.Println("uuid", moment.UUID)
		fmt.Println("uuid", moment.UserUUID)
		Expect("a030e570-ab6b-524d-815d-8aa23968a74e").To(Equal(moment.UUID))
	})

})
