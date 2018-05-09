package router

import (
	"net/http"
	"regexp"
	"strings"
)

// HTTPResponse is a response schema returned by subscribed function in case of HTTP event.
type HTTPResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func isHTTPEvent(r *http.Request) bool {
	// is request with custom event
	if r.Header.Get("event") != "" {
		return false
	}

	// is pre-flight CORS request with "event" header
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		corsReqHeaders := r.Header.Get("Access-Control-Request-Headers")
		headers := strings.Split(corsReqHeaders, ",")
		for _, header := range headers {
			if header == "event" {
				return false
			}
		}
	}

	return true
}

func extractPath(host, path string) string {
	extracted := path
	rxp, _ := regexp.Compile(hostedDomain)
	if rxp.MatchString(host) {
		subdomain := strings.Split(host, ".")[0]
		extracted = "/" + subdomain + path
	}
	return extracted
}
