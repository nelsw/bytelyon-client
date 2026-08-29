package main

import (
	"bytelyon-client/internal/app"
	"flag"
)

func main() {
	var level, url, key string
	flag.StringVar(&level, "--log", "debug", "log level trace->disabled")
	flag.StringVar(&url, "--url", "http://localhost:80", "web app api url")
	flag.StringVar(&key, "--key", "my-random-32-character-x-api-key", "web app api key")
	flag.Parse()
	app.New(url, key, level).Run()
}
