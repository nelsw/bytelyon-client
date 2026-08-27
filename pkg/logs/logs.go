package logs

import (
	"github.com/rs/zerolog/log"
)

func Init(args ...string) {
	log.Logger = MakeZerolog(args...)
}
