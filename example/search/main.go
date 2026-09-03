package main

import (
	"math/rand"

	"github.com/nelsw/bytelyon-client/internal/search"
)

func main() {

	search.Handle(
		rand.Intn(1_000),
		rand.Intn(10_000),
		false,
		"yoga mats",
		nil,
	)
}
