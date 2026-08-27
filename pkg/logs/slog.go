package logs

import (
	"log/slog"
	"os"

	slogzerolog "github.com/samber/slog-zerolog/v2"
)

func NewSlog() *slog.Logger {

	var sl slog.Level
	switch os.Getenv("SLOG_LEVEL") {
	case "debug":
		sl = slog.LevelDebug
	case "info":
		sl = slog.LevelInfo
	case "warn":
		sl = slog.LevelWarn
	case "error":
		sl = slog.LevelError
	default:
		sl = slog.LevelInfo
	}

	return slog.New(slogzerolog.Option{
		Level:  sl,
		Logger: NewZerolog(),
	}.NewZerologHandler())
}
