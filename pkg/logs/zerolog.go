package logs

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewZerolog() *zerolog.Logger {
	l := MakeZerolog()
	return &l
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
				return Cyan + l + Default
			case "DEB":
				return Purple + l + Default
			case "INF":
				return Green + l + Default
			case "WAR":
				return Yellow + l + Default
			case "ERR":
				return Red + l + Default
			case "FAT", "PAN":
				return RedBackground + White + l + Default
			default:
				return Default + l + Default
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

	if len(args) == 0 {
		s := os.Getenv("LOG_LEVEL")
		if s == "" {
			s = zerolog.InfoLevel.String()
		}
		args = append(args, s)
	}

	lvl, err := zerolog.ParseLevel(args[0])
	if err != nil {
		lvl = zerolog.TraceLevel
	}

	if l = l.Level(lvl); lvl == zerolog.TraceLevel {
		l = l.With().Caller().Logger()
	}

	return l
}
