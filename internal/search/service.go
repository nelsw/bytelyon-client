package search

import "github.com/nelsw/bytelyon-client/pkg/play"

func Fetch(
	botID int,
	searchID int,
	query string,
	headless bool,
) {
	play.Go(&Job{botID, searchID, headless, query})
}
