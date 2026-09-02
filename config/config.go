package config

import (
	"flag"
	"net/http"
	"os"

	"github.com/nelsw/bytelyon-client/pkg/file"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type value struct {
	Log  string `json:"log" yaml:"log"`
	Host string `json:"host" yaml:"host"`
	Tkn  string `json:"tkn" yaml:"tkn"`
	Dry  bool   `json:"dry" yaml:"dry"`
}

func (v value) ok() bool { return v.Host != "" && v.Tkn != "" }

var v value

func init() {

	/* System Env */
	v.Host = os.Getenv("API_HOST")
	v.Tkn = os.Getenv("API_TOKEN")
	v.Log = os.Getenv("APP_LOG")
	v.Dry = os.Getenv("DRY_RUN") == "true"
	if v.ok() {
		return
	}

	/* File Env */
	m, _ := godotenv.Read(".env")
	v.Host, _ = m["API_HOST"]
	v.Tkn, _ = m["API_TOKEN"]
	v.Log, _ = m["APP_LOG"]
	dryRun, _ := m["DRY_RUN"]
	v.Dry = dryRun == "true"
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
	flag.StringVar(&v.Log, "log", "debug", "log level trace->disabled")
	flag.StringVar(&v.Host, "host", "https://localhost", "web app api url")
	flag.StringVar(&v.Tkn, "tkn", "", "client api key")
	flag.BoolVar(&v.Dry, "dry", true, "dry run")
	flag.Parse()
}

func DryRun() bool            { return v.Dry }
func LogLvl() string          { return v.Log }
func ApiHost() string         { return v.Host }
func RunDry(b bool)           { v.Dry = b }
func AuthHeader() http.Header { return http.Header{"Authorization": []string{v.Tkn}} }
