package util

import (
	"net/http"
	"strings"
)

var routesToSkip = []string{"health", "debug", "metric"}

func IsFilteredHttpRoute(r *http.Request) bool {
	for _, route := range routesToSkip {
		if strings.Contains(r.RequestURI, route) {
			return true
		}
	}

	return false
}
