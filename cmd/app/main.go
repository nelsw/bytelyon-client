package main

import (
	"bytelyon-client/internal/job"
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func main() {
	var level, url, key string
	flag.StringVar(&level, "log", "debug", "log level trace->disabled")
	flag.StringVar(&url, "url", "http://localhost:80", "web app api url")
	flag.StringVar(&key, "key", "my-random-32-character-x-api-key", "web app api key")
	flag.Parse()

	api.Init(url, key)
	logs.Init(level)
	play.Init()

	log.Log().Msg(`🦁`)
	log.Log().Msg(`🦁  ByteLyon`)
	log.Log().Msg(`🦁`)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	q := job.StartQueue()
	for {
		select {
		case <-q.PollChan():
			q.Poll()
		case <-quit:
			q.Stop()
			return
		}
	}
}
