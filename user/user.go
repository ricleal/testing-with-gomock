package user

import "testing-with-gomock/doer"

type User struct {
	Doer doer.Doer
}

func (u *User) Use() error {
	return u.Doer.DoSomething(123, "Hello GoMock")
}

func NewUser(doer doer.Doer) *User {
	return &User{Doer: doer}
}
