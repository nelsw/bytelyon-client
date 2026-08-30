package main

import (
	"bytelyon-client/internal/config"
)

func main() {
	config.FromENV().Do()
}
