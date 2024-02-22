package url

import (
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"strings"
	"url-short/helpers"
)

// New creates a new instance of the API with logger and database number.
// It initializes a Redis client and sets up the API structure.
func New(logger *zerolog.Logger, dbNo int) *API {
	return &API{
		logger:     logger,
		repository: NewRedisClient(dbNo), // todo maybe redundant client idk
	}
}

// Shorten handles HTTP requests to shorten URLs.
// It decodes the JSON request body into a struct, validates the URL, checks request rate limits,
// and enforces URL formatting rules. If successful, it generates a short URL, stores it in the repository,
// and returns the shortened URL along with relevant metadata in the response.
// If any errors occur during processing, appropriate HTTP error responses are returned.
func (a *API) Shorten(w http.ResponseWriter, r *http.Request) {
	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, newErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	ip := getIPAddress(r)

	if !a.repository.isAllowed(ip) {
		// todo return custom error messages
		http.Error(w, newErrorResponse("Too Many Request"), http.StatusTooManyRequests)
		return
	}
	// todo add url validation with SuccessResponse code
	if !govalidator.IsURL(request.URL) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !helpers.RemoveDomainError(request.URL) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	request.URL = helpers.EnforceHTTP(request.URL)

	var shortenURL string
	if request.CustomShort == "" {
		shortenURL = uuid.New().String()[:6]
	} else {
		shortenURL = request.CustomShort
	}

	//	_, err = a.repository.Get(request.URL)
	//	if err != nil {
	//		// todo return error for already in use
	//		http.Error(w, err.Error(), http.StatusNotFound)
	//		return // return not found
	//	}

	err = a.repository.Set(shortenURL, request.URL, request.Expiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return // return 500
	}

	resp := SuccessResponse{
		URL:             os.Getenv("DOMAIN") + "/" + shortenURL,
		CustomShort:     "",
		Expiry:          request.Expiry,
		XRateRemaining:  0,
		XRateLimitReset: 30,
	}

	remainingRate, err := a.repository.calculateRemainingRate(ip)
	if err != nil {
		return
	}
	resp.XRateRemaining = remainingRate

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		//a.logger.Error().Str(l.KeyReqID, reqID).Err(err).Msg("")
		//e.ServerError(w, e.RespJSONEncodeFailure)
		return
	}

}

// Resolve handles HTTP requests to resolve shortened URLs.
// Retrieves the corresponding value from the repository,
// and redirects the client to the resolved URL using a 301 status code.
func (a *API) Resolve(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[1:]
	targetUrl, err := a.repository.Get(url)
	if err != nil {
		a.logger.Info().Str("info", url).Msg("message")
		http.Error(w, err.Error(), http.StatusNotFound) // todo show proper message
		return                                          // return not found
	}
	// todo add metric and logs
	http.Redirect(w, r, targetUrl, 301)
}

func getIPAddress(r *http.Request) string {
	// Get the remote address from the request
	remoteAddr := r.RemoteAddr
	// If the remote address contains a colon (for IPv6), we extract the IP portion
	if strings.Contains(remoteAddr, ":") {
		ipParts := strings.Split(remoteAddr, ":")
		remoteAddr = ipParts[0]
	}
	return remoteAddr
}
