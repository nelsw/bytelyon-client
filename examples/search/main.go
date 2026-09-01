package main

import "github.com/nelsw/bytelyon-client/internal/search"

func main() {

	err := search.Handle(1, 1, false, "ergonomic office chair", nil)
	if err != nil {
		panic(err)
	}

}
