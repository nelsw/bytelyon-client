package search

import (
	"path/filepath"
	"strconv"

	"github.com/nelsw/bytelyon-client/pkg/play"
	"github.com/rs/zerolog/log"
)

func Fetch(
	searchID int,
	query string,
	headless bool,
) {

	if err := play.Search(searchID, query, headless); err != nil {
		log.Warn().Err(err).Msg("failed to scrape page")
		return
	}

	path := filepath.Join(".storage", "search", strconv.Itoa(searchID), query)

	srcKey, imgKey, data := play.HandleFiles(path)

	UpdateSearch(searchID, imgKey, srcKey, data)
}
