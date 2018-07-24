// +build hosted

package router

import (
	"strings"
)

// extractPath extracts path from hosted EG host name (<space>.eventgateway([a-z-]*)?.io|slsgateway.com)
func extractPath(host, path string) string {
	subdomain := strings.Split(host, ".")[0]
	return basePath + subdomain + path
}

func systemPathFromSpace(space string) string {
	return basePath + space + "/"
}

// systemPathFromPath constructs path from path on which event was emitted. Helpful for "event.received" system event.
func systemPathFromPath(path string) string {
	return basePath + strings.Split(path, "/")[1] + "/"
}
