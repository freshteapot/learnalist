package alist_test

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing List types", func() {
	const (
		EmptyDataV1 = "v1-001.json"
		EmptyDataV2 = "v2-001.json"
		EmptyDataV3 = "v3-001.json"
		EmptyDataV4 = "v4-001.json"
		WithDataV1  = "v1-002.json"
		WithDataV2  = "v2-002.json"
		WithDataV3  = "v3-002.json"
		WithDataV4  = "v4-002.json"

		InvalidInteractSlideshowV1   = "v1-003.json"
		InvalidInteractTotalRecallV1 = "v1-004.json"
		InvalidInteractTotalRecallV2 = "v2-003.json"
	)

	It("Via New", func() {
		var tests = []struct {
			List       alist.Alist
			ExpectData string
			ExpectType string
		}{
			{
				List:       alist.NewTypeV1(),
				ExpectData: "TypeV1",
				ExpectType: alist.SimpleList,
			},
			{
				List:       alist.NewTypeV2(),
				ExpectData: "TypeV2",
				ExpectType: alist.FromToList,
			},
			{
				List:       alist.NewTypeV3(),
				ExpectData: "TypeV3",
				ExpectType: alist.Concept2,
			},
			{
				List:       alist.NewTypeV4(),
				ExpectData: "TypeV4",
				ExpectType: alist.ContentAndUrl,
			},
		}
		for _, test := range tests {
			Expect(test.List.Info.ListType).To(Equal(test.ExpectType))
			Expect(reflect.TypeOf(test.List.Data).Name()).To(Equal(test.ExpectData))
		}
	})

	It("Empty data via JSON", func() {
		var tests = []struct {
			Input      []byte
			ExpectType string
			ExpectSize func(aList alist.Alist)
		}{
			{
				Input:      testutils.GetTestData(EmptyDataV1),
				ExpectType: alist.SimpleList,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV1))).To(Equal(0))
				},
			},
			{
				Input:      testutils.GetTestData(EmptyDataV2),
				ExpectType: alist.FromToList,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV2))).To(Equal(0))
				},
			},
			{
				Input:      testutils.GetTestData(EmptyDataV3),
				ExpectType: alist.Concept2,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV3))).To(Equal(0))
				},
			},
			{
				Input:      testutils.GetTestData(EmptyDataV4),
				ExpectType: alist.ContentAndUrl,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV4))).To(Equal(0))
				},
			},
		}

		for _, test := range tests {
			var aList alist.Alist
			decoder := json.NewDecoder(bytes.NewReader(test.Input))
			err := decoder.Decode(&aList)
			Expect(err).To(BeNil())
			Expect(aList.Info.ListType).To(Equal(test.ExpectType))
			test.ExpectSize(aList)
		}
	})

	It("With data via JSON", func() {
		var tests = []struct {
			Input      []byte
			ExpectType string
			ExpectSize func(aList alist.Alist)
		}{
			{
				Input:      testutils.GetTestData(WithDataV1),
				ExpectType: alist.SimpleList,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV1))).To(Equal(7))
				},
			},
			{
				Input:      testutils.GetTestData(WithDataV2),
				ExpectType: alist.FromToList,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV2))).To(Equal(1))
				},
			},
			{
				Input:      testutils.GetTestData(WithDataV3),
				ExpectType: alist.Concept2,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV3))).To(Equal(1))
				},
			},
			{
				Input:      testutils.GetTestData(WithDataV4),
				ExpectType: alist.ContentAndUrl,
				ExpectSize: func(aList alist.Alist) {
					Expect(len(aList.Data.(alist.TypeV4))).To(Equal(2))
				},
			},
		}

		for _, test := range tests {
			var aList alist.Alist
			decoder := json.NewDecoder(bytes.NewReader(test.Input))
			err := decoder.Decode(&aList)
			Expect(err).To(BeNil())
			Expect(aList.Info.ListType).To(Equal(test.ExpectType))
			test.ExpectSize(aList)
		}
	})

	When("Interact", func() {
		It("invalid due to wrong value used for interact", func() {
			var tests = []struct {
				Input      []byte
				ExpectType string
			}{
				{
					Input:      testutils.GetTestData(InvalidInteractSlideshowV1),
					ExpectType: alist.SimpleList,
				},
				{
					Input:      testutils.GetTestData(InvalidInteractTotalRecallV1),
					ExpectType: alist.SimpleList,
				},
				{
					Input:      testutils.GetTestData(InvalidInteractTotalRecallV2),
					ExpectType: alist.FromToList,
				},
			}

			for _, test := range tests {
				var aList alist.Alist
				decoder := json.NewDecoder(bytes.NewReader(test.Input))
				err := decoder.Decode(&aList)
				Expect(err).To(BeNil())
				Expect(aList.Info.ListType).To(Equal(test.ExpectType))
				err = alist.Validate(aList)
				Expect(err).Should(HaveOccurred())
			}
		})
	})

	It("Make sure we reject empty lines", func() {

		var tests = []struct {
			List               alist.Alist
			ExpectErrorMessage string
		}{
			{
				List: func() alist.Alist {
					aList := alist.NewTypeV1()
					aList.Info.Title = "Hi"
					aList.Data = append(aList.Data.(alist.TypeV1), "")
					return aList
				}(),
				ExpectErrorMessage: i18n.ValidationAlistTypeV1,
			},
			{
				List: func() alist.Alist {
					aList := alist.NewTypeV2()
					aList.Info.Title = "Hi"
					aList.Data = append(aList.Data.(alist.TypeV2), alist.TypeV2Item{})
					return aList
				}(),

				ExpectErrorMessage: i18n.ValidationAlistTypeV2,
			},
			{
				List: func() alist.Alist {
					aList := alist.NewTypeV3()
					aList.Info.Title = "Hi"
					aList.Data = append(aList.Data.(alist.TypeV3), alist.TypeV3Item{})
					return aList
				}(),
				ExpectErrorMessage: i18n.ValidationAlistTypeV3,
			},
			{
				List: func() alist.Alist {
					aList := alist.NewTypeV4()
					aList.Info.Title = "Hi"
					aList.Data = append(aList.Data.(alist.TypeV4), alist.TypeV4Item{})
					return aList
				}(),
				ExpectErrorMessage: i18n.ValidationAlistTypeV4,
			},
		}

		for _, test := range tests {
			err := alist.Validate(test.List)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(test.ExpectErrorMessage))
		}
	})
})
