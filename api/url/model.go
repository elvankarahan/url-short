package url

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"time"
)

type API struct {
	logger     *zerolog.Logger
	repository *Repository
}

type Request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type SuccessResponse struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func newErrorResponse(error string) string {
	marshal, _ := json.Marshal(&ErrorResponse{Error: error})
	return string(marshal)
}
