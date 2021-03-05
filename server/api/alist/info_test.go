package alist_test

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing List Info", func() {
	const (
		KindDomainNotMatch = "info-from-001.json"
		ValidFrom          = "info-from-002.json"
		KindNotSupported   = "info-from-003.json"
		MissingRefUrl      = "info-from-004.json"
		MissingExtUUID     = "info-from-005.json"
		ValidFromCram      = "info-from-006.json"
	)

	FIt("Validating from", func() {
		var tests = []struct {
			Input       []byte
			ExpectError func(err error)
		}{
			{
				Input: testutils.GetTestData(ValidFrom),
				ExpectError: func(err error) {
					Expect(err).To(BeNil())
				},
			},
			{
				Input: testutils.GetTestData(KindNotSupported),
				ExpectError: func(err error) {
					Expect(err).To(Equal(i18n.ErrorInputSaveAlistFromKindNotSupported))
				},
			},
			{
				Input: testutils.GetTestData(MissingRefUrl),
				ExpectError: func(err error) {
					Expect(err).To(Equal(alist.ErrorListFromValid))
				},
			},
			{
				Input: testutils.GetTestData(MissingExtUUID),
				ExpectError: func(err error) {
					Expect(err).To(Equal(alist.ErrorListFromValid))
				},
			},
			{
				Input: testutils.GetTestData(KindDomainNotMatch),
				ExpectError: func(err error) {
					Expect(err).To(Equal(i18n.ErrorAListFromDomainMisMatch))
				},
			},
			{
				Input: testutils.GetTestData(ValidFromCram),
				ExpectError: func(err error) {
					Expect(err).To(BeNil())
				},
			},
		}

		for _, test := range tests {
			aList := alist.NewTypeV2()
			aList.Info.Title = "fake"

			var from openapi.AlistFrom

			json.Unmarshal(test.Input, &from)
			aList.Info.From = &from
			err := alist.Validate(aList)
			test.ExpectError(err)
		}
	})

	It("Ignore the from validation", func() {
		aList := alist.NewTypeV2()
		aList.Info.Title = "fake"
		aList.Info.SharedWith = aclKeys.SharedWithPublic

		err := alist.Validate(aList)
		Expect(err).To(BeNil())
	})

	When("Validating shared with when from is present", func() {
		var aList alist.Alist

		BeforeEach(func() {
			var from openapi.AlistFrom
			aList = alist.NewTypeV2()
			aList.Info.Title = "fake"
			fromInput := testutils.GetTestData(ValidFrom)
			json.Unmarshal(fromInput, &from)
			aList.Info.From = &from
		})

		When("Shared with public", func() {
			It("dont allow", func() {
				aList.Info.SharedWith = aclKeys.SharedWithPublic
				err := alist.Validate(aList)
				Expect(err).To(Equal(alist.ErrorSharingNotAllowedWithFrom))
			})

			It("unless learnalist ", func() {
				aList.Info.SharedWith = aclKeys.SharedWithPublic
				aList.Info.From.Kind = "learnalist"
				aList.Info.From.RefUrl = "http://learnalist.net/fake"
				err := alist.Validate(aList)
				Expect(err).To(BeNil())
			})
		})

		It("shared with friends", func() {
			aList.Info.SharedWith = aclKeys.SharedWithFriends
			err := alist.Validate(aList)
			Expect(err).To(Equal(alist.ErrorSharingNotAllowedWithFrom))
		})

		It("Not shared", func() {
			aList.Info.SharedWith = aclKeys.NotShared
			err := alist.Validate(aList)
			Expect(err).To(BeNil())
		})
	})

})
