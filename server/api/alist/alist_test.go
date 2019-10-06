package alist_test

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing alist.Alist", func() {
	When("Handling JSON", func() {
		Context("Unmarshal input", func() {
			It("Invalid json", func() {
				input := `{a}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("Failed to parse list."))
			})

			It("Missing info object", func() {
				input := `{"data": []}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("Failed to pass list. Info is missing."))
			})

			It("When info is not an object", func() {
				input := `{"info": ""}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("Failed to pass list. Something wrong with info object."))
			})

			It("When data is missing", func() {
				input := `{"info": {"title": "I am a title"}}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("Failed to pass list. Data is missing."))
			})
		})

		Context("Unmarshal invalid parsing of list types", func() {
			It("V1", func() {
				input := `{"data":[{}],"info":{"title":"I am a list","type":"v1"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(i18n.ValidationErrorListV1))
			})

			It("V2", func() {
				input := `{"data":[""],"info":{"title":"I am a list","type":"v2"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(i18n.ValidationErrorListV2))
			})

			It("V3", func() {
				input := `{"data":[""],"info":{"title":"I am a list","type":"v3"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(i18n.ValidationErrorListV3))
			})

			It("V4", func() {
				input := `{"data":[""],"info":{"title":"I am a list","type":"v4"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(i18n.ValidationErrorListV4))
			})

			It("Unsupported list tyype", func() {
				input := `{"data":[],"info":{"title":"I am a list","type":"na"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON([]byte(input))
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("Unsupported list type."))
			})
		})

		Context("Marshal input", func() {
			It("Valid json output based on json input", func() {
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON(validListTypeV1)
				Expect(err).ShouldNot(HaveOccurred())

				a, _ := json.Marshal(aList)
				Expect(a).To(Equal(validListTypeV1))
			})

			It("When the optional info.interact object is present", func() {
				aList := new(alist.Alist)
				err := aList.UnmarshalJSON(validListTypeV1WithInfoInteract)
				Expect(err).ShouldNot(HaveOccurred())

				a, _ := json.Marshal(aList)
				Expect(a).To(Equal(validListTypeV1WithInfoInteract))
			})
		})
	})
})
