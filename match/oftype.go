package match

import (
	"reflect"
)

type ofType struct {
	t string
}

func NewOfType(t string) *ofType {
	return &ofType{t}
}

func (o *ofType) Matches(x interface{}) bool {
	return reflect.TypeOf(x).String() == o.t
}

func (o *ofType) String() string {
	return "is of type " + o.t
}
