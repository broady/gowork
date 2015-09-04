package mypkg

import (
	"fmt"

	"golang.org/x/oauth2"
)

const Foo = "x"

func N() string {
	c := &oauth2.Config{}
	return fmt.Sprintf("%#v", c)
}
