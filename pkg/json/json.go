package json

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

func PrettyPrint(a any) {
	log.Print(string(Pretty(a)))
}

func Pretty(a any) []byte {
	b, err := json.MarshalIndent(&a, "", "\t")
	if err != nil {
		log.Warn().Err(err).Msg("cannot pretty print")
	}
	return b
}

func Marshal(a any) (b []byte, err error) {
	if b, err = json.Marshal(&a); err != nil {
		log.Warn().Err(err).Msg("cannot marshal")
	}
	return
}

func Unmarshal(b []byte, a any) (err error) {
	if err = json.Unmarshal(b, &a); err != nil {
		log.Warn().Err(err).Msg("cannot unmarshal")
	}
	return
}
