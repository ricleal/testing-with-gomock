package doer

import (
	"fmt"
	"strconv"
)

//go:generate mockgen -destination=../mocks/mock_doer.go -package=mocks testing-with-gomock/doer Doer

type Doer interface {
	DoSomething(int, string) error
}

type doerImpl struct{}

func NewDoer() *doerImpl {
	return &doerImpl{}
}

func (d *doerImpl) DoSomething(n int, str string) error {
	fmt.Printf("DoSomething %s %d times\n", str, n)
	if strconv.Itoa(n) == str {
		return nil
	}
	return fmt.Errorf("error: %s != %d", str, n)
}
