package alist_test

import (
	"reflect"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing List type V2", func() {
	It("Via New", func() {
		aList := alist.NewTypeV2()
		Expect(aList.Info.ListType).To(Equal(alist.FromToList))
		Expect(reflect.TypeOf(aList.Data).Name()).To(Equal("TypeV2"))
	})
})
