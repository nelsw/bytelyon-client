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
	Reset             = "\033[0m"
	Default           = Reset
	Red               = "\033[0;31m"
	Green             = "\033[0;32m"
	Purple            = "\033[0;35m"
	Cyan              = "\033[0;36m"
	White             = "\033[0;37m"
	RedBackground     = "\033[41m"
	YellowBoldIntense = "\033[1;93m"
	BlueBoldIntense   = "\033[1;94m"
	WhiteBoldIntense  = "\033[1;97m"
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
		TimeFormat: time.TimeOnly,
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
				return WhiteBoldIntense + " 🦁" + Default
			case "TRA":
				return Cyan + l + Default
			case "DEB":
				return Purple + l + Default
			case "INF":
				return Green + l + Default
			case "WAR":
				return "\033[0;5m" + l + Default
			case "ERR":
				return Red + l + Default
			case "FAT", "PAN":
				return RedBackground + White + l + Default
			default:
				return Default + l + Default
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

func PrintBanner() {
	println(strings.Join([]string{
		"\n",
		YellowBoldIntense + `* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
		YellowBoldIntense + `*                                                                           * `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*` + BlueBoldIntense + `    ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝    ` + YellowBoldIntense + `* `,
		YellowBoldIntense + `*                                                                           * `,
		YellowBoldIntense + `* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * `,
		"\n",
	}, "\n") + Reset)
}
