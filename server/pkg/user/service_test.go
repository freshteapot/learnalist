package user_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tideland/gorest/jwt"
)

var _ = Describe("Testing User from IDP", func() {
	FIt("", func() {
		// "github.com/Timothylock/go-signin-with-apple/apple"
		idpToken := `XXX`

		//var target jwt.Claims
		//jwt.Decode(idpToken, target)
		j, err := jwt.Decode(idpToken)
		fmt.Println(err)

		leeway := time.Hour
		fmt.Println(j.IsValid(leeway))
		fmt.Println(j.Claims().Get("aud"))
		fmt.Println(j.Claims().Get("iss"))
		Expect("a").To(Equal("a"))
	})
})
