package main

import (
	"math/rand"

	"github.com/nelsw/bytelyon-client/internal/sitemap"
)

func main() {
	sitemap.Handle(rand.Intn(10), rand.Intn(100), false, "bytelyon.com")
}
