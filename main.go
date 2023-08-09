package main

import (
	"fmt"
	"testing-with-gomock/doer"
	"testing-with-gomock/user"
)

func main() {
	fmt.Println("start")

	doer := doer.NewDoer()
	user := user.NewUser(doer)

	err := user.Use()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("end")
}
