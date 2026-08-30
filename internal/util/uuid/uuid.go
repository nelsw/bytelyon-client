package uuid

import "github.com/google/uuid"

func FromURL(url string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
}
