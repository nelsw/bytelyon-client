package scrape

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nelsw/bytelyon-client/pkg/uuid"
)

func Page(headless bool, url string, dirs ...any) (key string, err error) {

	paths := []string{".storage"}
	for _, dir := range dirs {
		paths = append(paths, fmt.Sprintf("%v", dir))
	}

	path := filepath.Join(paths...)
	name := uuid.FromURL(url).String()

	key = filepath.Join(path, name)
	err = exec.Command(
		"./scripts/async_page",
		"-u", url,
		"-p", path,
		"-n", name,
		"-m", strconv.FormatBool(headless),
	).Run()

	return
}

func Serp(headless bool, query string, dirs ...any) (key string, err error) {

	paths := []string{".storage"}
	for _, dir := range dirs {
		paths = append(paths, fmt.Sprintf("%v", dir))
	}

	path := filepath.Join(paths...)
	name := uuid.FromURL("https://wwww.google.com?q=" + strings.ReplaceAll(query, " ", "+")).String()

	key = filepath.Join(path, name)
	err = exec.Command(
		"./scripts/sync_search",
		"-q", query,
		"-p", path,
		"-n", name,
		"-m", strconv.FormatBool(headless),
	).Run()

	return
}
