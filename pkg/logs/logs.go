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

func init() {
	Init()
}

func Init() {
	if initd {
		return
	}
	initd = true
	var err error
	if f, err = os.Create(time.Now().UTC().Format(time.DateOnly) + ".log"); err != nil {
		panic(err)
	}
	log.Logger = MakeZerolog()
}

func MakeZerolog() zerolog.Logger {

	lvl, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
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

func NewSlog() *slog.Logger {

	var sl slog.Level
	switch os.Getenv("LOG_LEVEL") {
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

func Banner() {
	println(strings.Join([]string{
		"\n\u001B[1;93m",
		`* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
		`*                                                                           * `,
		`*[1;94m` + `    ██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗    ` + "\033[1;93m" + `* `,
		`*[1;94m` + `    ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║    ` + "\033[1;93m" + `* `,
		`*[1;94m` + `    ██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║    ` + "\033[1;93m" + `* `,
		`*[1;94m` + `    ██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║    ` + "\033[1;93m" + `* `,
		`*[1;94m` + `    ██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║    ` + "\033[1;93m" + `* `,
		`*[1;94m` + `    ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝    ` + "\033[1;93m" + `* `,
		`*                                                                           * `,
		`* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
		"\n\u001B[0m",
	}, "\n"))
}
