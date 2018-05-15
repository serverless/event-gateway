package http

import (
	httppkg "net/http"
	"strings"
)

// FlattenHeader takes http.Header and flatten value array
// (map[string][]string -> map[string]string) so it's easier
// to access headers by user.
func FlattenHeader(req httppkg.Header) map[string]string {
	headers := map[string]string{}
	for key, header := range req {
		headers[strings.ToLower(key)] = header[0]
		if len(header) > 1 {
			headers[strings.ToLower(key)] = strings.Join(header, ", ")
		}
	}

	return headers
}
