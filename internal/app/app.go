package app

import (
	"bytelyon-client/internal/job"
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

type App struct {
	jobs *job.Queue
	quit chan os.Signal
}

func New(url, key, level string) *App {

	api.Init(url, key)
	logs.Init(level)
	play.Init()

	a := &App{
		quit: make(chan os.Signal, 1),
		jobs: job.NewQueue(),
	}
	signal.Notify(a.quit, syscall.SIGINT, syscall.SIGTERM)

	log.Log().Msg(`🦁`)
	log.Log().Msg(`🦁  ByteLyon Client App`)
	log.Log().Msgf("🦁\n")

	return a
}

func (a *App) Run() {
	a.jobs.Start()
	for {
		select {
		case <-a.jobs.PollInterval():
			a.jobs.Poll()
		case <-a.quit:
			a.jobs.Stop()
			log.Log().Msgf("\n👋\n")
			return
		}
	}
}

func (a *App) Quit() { a.quit <- syscall.SIGINT }
