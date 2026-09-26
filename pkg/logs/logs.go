package logs

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	Reset   = "\033[0m"
	Default = Reset
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
	Purple  = "\033[0;35m"
	Cyan    = "\033[0;36m"
)

var f *os.File
var initd bool

func Init() {
	if initd {
		return
	}
	initd = true
	const dir = ".storage/logs/"
	var err error
	if err = os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	} else if f, err = os.Create(dir + time.Now().UTC().Format(time.DateOnly) + ".log"); err != nil {
		panic(err)
	}
	log.Logger = MakeZerolog()
}

func MakeZerolog() zerolog.Logger {

	lvl, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil || lvl == zerolog.NoLevel { // an empty LOG_LEVEL parses to NoLevel, which hides everything
		lvl = zerolog.TraceLevel
	}

	logger := zerolog.New(zerolog.MultiLevelWriter(f, zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
		FieldsOrder: []string{
			"#", "id",
			"t", "type",
			"q", "query",
		},
		FormatLevel: func(a any) string {
			if a == nil || a == "<nil>" {
				a = "NIL"
			}
			switch l := strings.ToUpper(a.(string)[:3]); l {
			case "NIL":
				return Default + " 🦁" + Default
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
				return "\033[41m\u001B[0;37m" + l + Default
			default:
				return Default + l + Default
			}
		},
	})).Level(lvl).With().Timestamp().Logger()

	if lvl == zerolog.TraceLevel {
		return logger.With().Caller().Logger()
	}

	return logger
}
