package main

import (
	"math/rand"

	"github.com/nelsw/bytelyon-client/internal/search"
)

func main() {

	err := search.Handle(
		rand.Intn(1_000),
		rand.Intn(10_000),
		false,
		"fire blanket for sale",
		nil,
	)
	if err != nil {
		panic(err)
	}

}
