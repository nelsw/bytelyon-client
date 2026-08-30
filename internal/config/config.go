package config

import (
	"bytelyon-client/internal/model"
	"bytelyon-client/internal/provider/api"
	"bytelyon-client/internal/provider/logs"
	"bytelyon-client/internal/provider/play"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	YellowBoldIntense = "\033[1;93m"
	BlueBoldIntense   = "\033[1;94m"
)

var lvl, url, key string

func FromCLI() {
	flag.StringVar(&lvl, "log", "debug", "log level trace->disabled")
	flag.StringVar(&url, "url", "https://bytelyon.com", "web app api url")
	flag.StringVar(&key, "key", "", "client api key")
	flag.Parse()
	Init()
	printBanner()
}

func FromENV() *model.Bot {

	var profile string
	flag.StringVar(&profile, "profile", "", "env profile to use, eg env.testing")
	flag.Parse()

	var env map[string]string
	var err error

	if profile != "" {
		env, err = godotenv.Read(".env." + profile)
	} else {
		env, err = godotenv.Read(".env")
	}
	fmt.Println(env)
	if err != nil {
		panic(err)
	}

	lvl = env["LOG_LEVEL"]
	url = env["API_URL"]
	key = env["API_KEY"]

	Init()

	var bot model.Bot
	if s, ok := env["BOT_ID"]; ok {
		bot.ID, _ = strconv.Atoi(s)
	}
	if s, ok := env["BOT_CHILD_ID"]; ok {
		bot.ChildID, _ = strconv.Atoi(s)
	}
	if s, ok := env["BOT_TYPE"]; ok {
		bot.Type = model.BotType(s)
	}
	if s, ok := env["BOT_QUERY"]; ok {
		bot.Query = s
	}
	if s, ok := env["BOT_HEADLESS"]; ok {
		bot.Headless, _ = strconv.ParseBool(s)
	}
	if s, ok := env["BOT_BLACKLIST"]; ok {
		bot.Blacklist = make(map[string]bool)
		for k := range strings.SplitSeq(s, "\n") {
			bot.Blacklist[k] = true
		}
	}
	if s, ok := env["BOT_PLAYED_AT"]; ok {
		bot.PlayedAt, _ = time.Parse(time.RFC3339, s)
	}

	printBanner()

	return &bot
}

func Init() {
	logs.Init(lvl)
	api.Init(url, key)
	play.Init()
}

func printBanner() {
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
	}, "\n") + "\033[0m")
}
