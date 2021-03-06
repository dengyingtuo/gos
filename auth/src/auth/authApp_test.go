package main

import (
	"account"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"goslib/redisdb"
)

var _ = Describe("AuthApp", func() {
	redisdb.Connect("localhost:6379", "", 0)

	It("should startup", func() {
		accountId := "fakeAccountId"
		password := "fakePassword"

		account.Delete(accountId)
		account.Create(accountId, password)

		user, _ := account.Lookup(accountId)
		fmt.Println("Found user: ", user.Username)
		Expect(user).ToNot(BeNil())
	})
})
