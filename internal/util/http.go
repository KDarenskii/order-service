package util

import (
	"net/http"
	"strings"
)

var routesToSkip = []string{"/health", "/debug", "/metrics"}

func IsFilteredHttpRoute(r *http.Request) bool {
	path := r.URL.Path

	for _, prefix := range routesToSkip {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}
