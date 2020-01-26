package e2e_test

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
)

var usernameOwner = "iamchris"
var password = "test123"
var usernameReader = "iamusera"

var server = "http://127.0.0.1:1234"
var inputAlistV1 = `
{
  "data": [
      "monday",
      "tuesday",
      "wednesday",
      "thursday",
      "friday",
      "saturday",
      "sunday"
  ],
  "info": {
      "title": "Days of the Week",
      "type": "v1",
			"shared_with": "%s"
  }
}
`

var inputAlistV2 = `
{
  "data": [
  {
    "from":"car",
    "to": "bil"
  }
  ],
  "info": {
  	"title": "Days of the Week",
  	"type": "v2",
  	"shared_with": "%s",
  	"labels": [
    	"water"
  	]
	}
}`

var inputAlistV3 = `{
  "info": {
      "title": "Getting my row on.",
      "type": "v3",
      "shared_with": "%s"
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
}`

var inputAlistV4 = `
{
  "info": {
      "title": "A list of fine quotes.",
      "type": "v4",
      "shared_with": "%s"
  },
  "data": [
    {
      "content": "Im tough, Im ambitious, and I know exactly what I want. If that makes me a bitch, okay. ― Madonna",
      "url":  "https://www.goodreads.com/quotes/54377-i-m-tough-i-m-ambitious-and-i-know-exactly-what-i"
    },
    {
      "content": "Design is the art of arranging code to work today, and be changeable forever. – Sandi Metz",
      "url":  "https://dave.cheney.net/paste/clear-is-better-than-clever.pdf"
    }
  ]
}
`
var usernameIncrement int64

func getInputListWithShare(listType string, sharedWith string) string {
	var inputAlist string

	switch listType {
	case alist.SimpleList:
		inputAlist = inputAlistV1
	case alist.FromToList:
		inputAlist = inputAlistV2
	case alist.Concept2:
		inputAlist = inputAlistV3
	case alist.ContentAndUrl:
		inputAlist = inputAlistV4
	default:
		panic("List input type not supported")
	}

	with := ""
	switch sharedWith {
	case aclKeys.SharedWithPublic:
		with = aclKeys.SharedWithPublic
	case aclKeys.SharedWithFriends:
		with = aclKeys.SharedWithFriends
	case aclKeys.NotShared:
		with = aclKeys.NotShared
	}
	return fmt.Sprintf(inputAlist, with)
}

func generateUsername() string {

	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()

	return usernameOwner + "-" + str
}

func cleanEchoJSONResponse(data []byte) string {
	return strings.TrimSuffix(string(data), "\n")
}
