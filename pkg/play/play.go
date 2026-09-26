package play

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nelsw/bytelyon-client/internal/bot"
	"github.com/nelsw/bytelyon-client/pkg/fs"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog/log"
)

const (
	news    = "./scripts/news"
	search  = "./scripts/sync_search"
	sitemap = "./scripts/pages"
)

func run(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Pages(t bot.Type, id int, headless bool, urls []string) error {
	return run(sitemap, append([]string{
		"-t", string(t),
		"-i", strconv.Itoa(id),
		"-m", strconv.FormatBool(headless),
		"-u",
	}, urls...))
}

func News(headless bool, urls []string) error {
	return run(news, append([]string{
		"-m", strconv.FormatBool(headless),
		"-u",
	}, urls...))
}

func Search(id int, query string, headless bool) error {
	return run(search, []string{
		"-i", strconv.Itoa(id),
		"-q", query,
		"-m", strconv.FormatBool(headless),
	})
}

func HandleFiles(path string) (srcKey string, imgKey string, data model.Data[string, any]) {

	data = model.Data[string, any]{}

	from := path + ".html"
	srcKey = strings.ReplaceAll(from, ".storage/", "")
	_ = fs.Move(from, srcKey)
	log.Trace().Msg("moved html file")

	from = path + ".png"
	imgKey = strings.ReplaceAll(from, ".storage/", "")
	_ = fs.Move(from, imgKey)
	log.Trace().Msg("moved image file")

	if bytes, err := os.ReadFile(path + ".json"); err != nil {
		log.Warn().Err(err).Msg("failed to read serp")
	} else if err = json.Unmarshal(bytes, &data); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal serp")
	} else {
		log.Trace().Msg("unmarshaled data")
	}
	return
}
