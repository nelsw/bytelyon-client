package logs

import (
	"log/slog"
	"os"
	"strings"
	"time"

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

func init() {
	Init(zerolog.TraceLevel.String())
}

func Init(args ...string) {
	log.Logger = MakeZerolog(args...)
}

func MakeZerolog(args ...string) zerolog.Logger {

	var lvlStr string
	if len(args) > 0 {
		lvlStr = args[0]
	}

	lvl, err := zerolog.ParseLevel(lvlStr)
	if err != nil {
		lvl = zerolog.TraceLevel
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
		FieldsOrder: []string{
			"#", "id",
			"t", "type",
			"q", "query",
		},
		FormatLevel: func(a any) string {
			if a == nil || a == "<nil>" {
				a = "   "
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "TRA":
				return "\033[0;36m" + l + "\033[0m"
			case "DEB":
				return "\033[0;35m" + l + "\033[0m"
			case "INF":
				return "\033[0;32m" + l + "\033[0m"
			case "WAR":
				return "\033[0;33m" + l + "\033[0m"
			case "ERR":
				return "\033[0;31m" + l + "\033[0m"
			case "FAT", "PAN":
				return "\033[41m" + "\033[0;37m" + l + "\033[0m"
			default:
				return ""
			}
		},
	}).Level(lvl).With().Timestamp().Logger()

	if lvl == zerolog.TraceLevel {
		return logger.With().Caller().Logger()
	}

	return logger
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
