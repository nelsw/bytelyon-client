package logs

import (
	"log/slog"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[0;33m"
	Purple = "\033[0;35m"
	Cyan   = "\033[0;36m"
	White  = "\033[0;37m"
	RedBG  = "\033[41m"
)

func Init(args ...string) {
	log.Logger = MakeZerolog(args...)
}

func MakeZerolog(args ...string) zerolog.Logger {

	l := log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatLevel: func(a any) string {
			if a == nil {
				a = "info"
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return Cyan + l + Reset
			case "DEB":
				return Purple + l + Reset
			case "INF":
				return Green + l + Reset
			case "WAR":
				return Yellow + l + Reset
			case "ERR":
				return Red + l + Reset
			case "FAT", "PAN":
				return RedBG + White + l + Reset
			default:
				return Reset + l + Reset
			}
		},
		FieldsOrder: []string{
			"request", "response",
			"ƒ",
			"ready",
			"userId", "botId", "id",
			"type", "botType",
			"target",
			"size",
			"table",
			"domain",
			"method", "authorization", "path", "query",
			"code",
			"body",
			"isAuthorized", "context",
		},
	})

	var lvlStr string
	if len(args) > 0 {
		lvlStr = args[0]
	}

	lvl, err := zerolog.ParseLevel(lvlStr)
	if err != nil {
		lvl = zerolog.DebugLevel
	}

	if l = l.Level(lvl); lvl == zerolog.TraceLevel {
		l = l.With().Caller().Logger()
	}

	return l
}

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
		Logger: new(MakeZerolog()),
	}.NewZerologHandler())
}
