package helpers

import (
	"os"
	"strings"
)

// EnforceHTTP ensures that the given URL has an HTTP scheme.
// If the URL does not start with "http", it prepends "http://" to the URL.
// It then returns the modified URL.
func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

// RemoveDomainError checks if the given URL matches the domain specified in the environment variable.
// If the URL matches the domain, it returns false, indicating no error.
// Otherwise, it removes common URL prefixes (http://, https://, www.) and extracts the domain.
// It then checks if the extracted domain matches the domain environment.
func RemoveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	return newURL != os.Getenv("DOMAIN")
}
