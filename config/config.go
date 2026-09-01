package config

import (
	"flag"
	"os"

	"github.com/nelsw/bytelyon-client/pkg/file"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type value struct {
	AppLog string `json:"app_log" yaml:"appLog"`
	ApiURL string `json:"api_url" yaml:"apiURL"`
	ApiKey string `json:"api_key" yaml:"apiKey"`
	DryRun bool   `json:"dry_run" yaml:"dryRun"`
}

func (v value) ok() bool { return v.ApiURL != "" && v.ApiKey != "" }

var v value

func init() {

	/* System Env */
	v.ApiURL = os.Getenv("API_URL")
	v.ApiKey = os.Getenv("API_KEY")
	v.AppLog = os.Getenv("APP_LOG")
	v.DryRun = os.Getenv("DRY_RUN") == "true"
	if v.ok() {
		return
	}

	/* File Env */
	m, _ := godotenv.Read(".env")
	v.ApiURL, _ = m["API_URL"]
	v.ApiKey, _ = m["API_KEY"]
	v.AppLog, _ = m["APP_LOG"]
	dryRun, _ := m["DRY_RUN"]
	v.DryRun = dryRun == "true"
	if v.ok() {
		return
	}

	/* File JSON */
	if v, _ = file.ReadType[value](".json", true); v.ok() {
		return
	}

	/* File Y(A)ML */
	if s, err := file.ReadStr(".yaml", true); err == nil {
		if _ = yaml.Unmarshal([]byte(os.ExpandEnv(s)), &v); v.ok() {
			return
		}
	}

	/* CLI Args */
	flag.StringVar(&v.AppLog, "log", "debug", "log level trace->disabled")
	flag.StringVar(&v.ApiURL, "url", "https://localhost", "web app api url")
	flag.StringVar(&v.ApiKey, "key", "", "client api key")
	flag.BoolVar(&v.DryRun, "dry-run", false, "dry run")
	flag.Parse()
}

func ApiURL() string { return v.ApiURL }
func ApiKey() string { return v.ApiKey }
func DryRun() bool   { return v.DryRun }
func LogLvl() string { return v.AppLog }
