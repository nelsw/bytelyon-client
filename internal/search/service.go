package search

import "github.com/nelsw/bytelyon-client/pkg/play"

func Fetch(
	searchID int,
	query string,
	headless bool,
) {
	play.Go(&Job{searchID, headless, query})
}
