package play

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/nelsw/bytelyon-client/internal/bot"
)

func Pages(t bot.Type, id int, headless bool, urls []string) error {

	// collect our initial args
	args := []string{
		"-t", t.String(),
		"-i", strconv.Itoa(id),
		"-m", strconv.FormatBool(headless),
		"-u",
	}

	// pass each url as an arg
	args = append(args, urls...)

	return run("./scripts/pages", args)
}

func News(headless bool, urls []string) error {

	// collect our initial args
	args := []string{
		"-m", strconv.FormatBool(headless),
		"-u",
	}

	// pass each url as an arg
	args = append(args, urls...)

	return run("./scripts/news", args)
}

func Search(id int, query string, headless bool) error {
	return run("./scripts/sync_search", []string{
		"-i", strconv.Itoa(id),
		"-q", query,
		"-m", strconv.FormatBool(headless),
	})
}

func run(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
