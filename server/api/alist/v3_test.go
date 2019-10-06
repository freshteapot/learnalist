package alist_test

import (
	"reflect"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing List type V3", func() {
	It("Via New", func() {
		aList := alist.NewTypeV3()
		Expect(aList.Info.ListType).To(Equal(alist.Concept2))
		Expect(reflect.TypeOf(aList.Data).Name()).To(Equal("TypeV3"))

		Expect(len(aList.Info.Labels)).To(Equal(2))
		Expect(aList.Info.Labels[0]).To(Equal("rowing"))
		Expect(aList.Info.Labels[1]).To(Equal("concept2"))
	})

	When("Checking the data structure", func() {
		Context("Json", func() {
			var input string
			BeforeEach(func() {
				input = `[{
					"when": "2019-05-06",
					"overall": {
						"time": "7:15.9",
						"distance": 2000,
						"spm": 28,
						"p500": "1:48.9"
					},
					"splits": [
						{
							"time": "1:46.4",
							"distance": 500,
							"spm": 29,
							"p500": "1:58.0"
						}
					]
				}]`
			})
			It("Parse", func() {
				_, err := alist.ParseTypeV3([]byte(input))
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("Validate", func() {
				data, err := alist.ParseTypeV3([]byte(input))
				Expect(err).ShouldNot(HaveOccurred())
				err = alist.ValidateTypeV3(data)
				Expect(err).ShouldNot(HaveOccurred())
			})

			Context("Reject just enough to check the logic on ValidateTypeV3", func() {
				var data alist.TypeV3
				var err error
				BeforeEach(func() {
					data, err = alist.ParseTypeV3([]byte(input))
					Expect(err).ShouldNot(HaveOccurred())
				})

				It("When", func() {
					data[0].When = ""
					err = alist.ValidateTypeV3(data)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal(i18n.ValidationAlistTypeV3))
				})

				It("Overall", func() {
					data[0].Overall.Distance = 0
					err = alist.ValidateTypeV3(data)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal(i18n.ValidationAlistTypeV3))
				})

				It("A bad Split", func() {
					data[0].Splits[0].Distance = 0
					err = alist.ValidateTypeV3(data)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal(i18n.ValidationAlistTypeV3))
				})
			})
		})

		Context("TypeV3Item", func() {
			Context("Distance", func() {
				It("Valid", func() {
					input := 2000
					err := alist.ValidateTypeV3Distance(input)
					Expect(err).ShouldNot(HaveOccurred())
				})

				It("Not valid because it is zero", func() {
					input := 0
					err := alist.ValidateTypeV3Distance(input)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Should not be empty."))
				})
			})
			Context("When", func() {
				It("Valid", func() {
					input := "2019-05-15"
					err := alist.ValidateTypeV3When(input)
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("The day must have leading 0 if under 10", func() {
					input := "2019-05-5"
					err := alist.ValidateTypeV3When(input)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("When should be YYYY-MM-DD."))
				})

				It("The date cant be seperated with /", func() {
					input := "2019/05/01"
					err := alist.ValidateTypeV3When(input)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("When should be YYYY-MM-DD."))
				})
				It("The date is empty", func() {
					input := ""
					err := alist.ValidateTypeV3When(input)
					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("When should be YYYY-MM-DD."))
				})
			})
			/*
			   type V3Split struct {
			   	Time     string `json:"time"`
			   	Distance int    `json:"distance"`
			   	Spm      int    `json:"spm"`
			   	P500     string `json:"p500"`
			   }
			*/
			Context("Overall, a single split", func() {
				var input alist.V3Split
				BeforeEach(func() {
					input = alist.V3Split{
						Time:     "7:15.9",
						Distance: 2000,
						Spm:      28,
						P500:     "1:48.9",
					}
				})
				Context("V3Split", func() {
					It("Valid", func() {
						err := alist.ValidateTypeV3Split(input)
						Expect(err).ShouldNot(HaveOccurred())
					})

					Context("Time", func() {
						It("Missing the minute or hour : split", func() {
							input.Time = "49.9"
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Is not valid format."))
						})
						It("Wrong format as it is lacking the : and too many .", func() {
							input.Time = "1.0.1"
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Is not valid format."))
						})

						It("Wrong format as it is has two :.", func() {
							input.Time = "1:00:0"
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Is not valid format."))
						})

						It("Not valid due to time being empty", func() {
							input.Time = ""
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Should not be empty."))
						})
					})
					Context("Distance", func() {
						It("Valid", func() {
							input.Distance = 2000
							err := alist.ValidateTypeV3Split(input)
							Expect(err).ShouldNot(HaveOccurred())
						})
						It("Not valid", func() {
							input.Distance = 0
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Should not be empty."))
						})
					})

					Context("Spm", func() {
						It("Valid", func() {
							input.Spm = 29
							err := alist.ValidateTypeV3Split(input)
							Expect(err).ShouldNot(HaveOccurred())
						})
						It("Too low", func() {
							input.Spm = 9
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Stroke per minute should be between the range 10 and 50."))
						})
						It("Too high", func() {
							input.Spm = 51
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Stroke per minute should be between the range 10 and 50."))
						})
					})
					Context("P500", func() {
						// The invalid are covered by ValidateTypeV3Time.
						It("Valid time for P500", func() {
							input.P500 = "1:49.9"
							err := alist.ValidateTypeV3Split(input)
							Expect(err).ShouldNot(HaveOccurred())
						})

						It("Not valid", func() {
							input.P500 = ""
							err := alist.ValidateTypeV3Split(input)
							Expect(err).Should(HaveOccurred())
							Expect(err.Error()).To(Equal("Should not be empty."))
						})
					})
				})
			})
		})
	})
})

/*
package alist

import (
	"reflect"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/stretchr/testify/assert"
)

func TestNewTypeV3(t *testing.T) {
	aList := NewTypeV3()
	assert.Equal(t, Concept2, aList.Info.ListType)
	assert.Equal(t, "TypeV3", reflect.TypeOf(aList.Data).Name())
	assert.Equal(t, 2, len(aList.Info.Labels))
	assert.Equal(t, "rowing", aList.Info.Labels[0])
	assert.Equal(t, "concept2", aList.Info.Labels[1])
}

func TestAlistTypeV3(t *testing.T) {
	input := `
{
  "info": {
      "title": "Getting my row on.",
      "type": "v3"
  },
  "data": [{
    "when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": 2000,
      "spm": 28,
      "p500": "1:48.9"
    },
    "splits": [
      {
        "time": "1:46.4",
        "distance": 500,
        "spm": 29,
        "p500": "1:58.0"
      }
    ]
  }]
}
`
	jsonBytes := []byte(input)
	aList := new(Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	assert.Nil(t, err)
	assert.Equal(t, "2019-05-06", aList.Data.(TypeV3)[0].When)
	// Confirm the enrichment happened.
	assert.Equal(t, 2, len(aList.Info.Labels))

	err = ValidateTypeV3(*aList)
	assert.Nil(t, err)

	typeV3Item := aList.Data.(TypeV3)[0]
	typeV3Item.When = ""
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.When = "2019-05-06"

	typeV3Item.Overall.Distance = 0
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Distance = 2000

	typeV3Item.Overall.Spm = 9
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Spm = 28

	typeV3Item.Overall.Time = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Time = "7:15.9"

	typeV3Item.Overall.P500 = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.P500 = "1:10.0"

	after := enrichTypeV3(*aList)
	assert.Equal(t, 2, len(after.Info.Labels))
	// Make sure we dont duplicate the labels
	after = enrichTypeV3(after)
	assert.Equal(t, 2, len(after.Info.Labels))

	// Test a bad split
	typeV3Item.Splits[0].Time = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = ValidateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Splits[0].Time = "1:0.0"
}

func TestTypeV3(t *testing.T) {
	input := `[{
    "when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": 2000,
      "spm": 28,
      "p500": "1:48.9"
    },
    "splits": [
      {
        "time": "1.46.4",
        "distance": 500,
        "spm": 29,
        "p500": "1:58.0"
      }
    ]
  }]
`
	jsonBytes := []byte(input)
	_, err := ParseTypeV3(jsonBytes)
	assert.Nil(t, err)
}
*/
