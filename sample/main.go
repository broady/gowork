package main

import (
	"log"

	"./mypkg"

	"github.com/satori/go.uuid"
)

func main() {
	log.Print(mypkg.Foo)
	log.Print(mypkg.N())
	log.Print(uuid.NewV4())
}
