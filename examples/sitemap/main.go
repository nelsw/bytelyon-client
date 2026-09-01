package main

import "github.com/nelsw/bytelyon-client/internal/sitemap"

func main() {
	err := sitemap.Handle(1, 1, false, "publix.com")
	if err != nil {
		panic(err)
	}
}
