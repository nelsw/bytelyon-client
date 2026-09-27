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

func run(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Pages(t bot.Type, id int, headless bool, urls []string) error {
	return run("./scripts/pages", append([]string{
		"-t", string(t),
		"-i", strconv.Itoa(id),
		"-m", strconv.FormatBool(headless),
		"-u",
	}, urls...))
}

func Search(id int, query string, headless bool) error {
	return run("./scripts/sync_search", []string{
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

	from = path + ".json"
	if bytes, err := os.ReadFile(from); err != nil {
		log.Warn().Err(err).Msg("failed to read file")
	} else if err = json.Unmarshal(bytes, &data); err != nil {
		log.Warn().Err(err).Msg("failed to unmarshal file")
	} else {
		log.Trace().Msg("unmarshaled data")
		_ = os.Remove(from)
	}
	return
}
