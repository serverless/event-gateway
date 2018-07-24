// +build hosted

package router

import (
	"regexp"
	"strings"
)

var hostedDomainPattern *regexp.Regexp

func init() {
	hostedDomainPattern = regexp.MustCompile("(eventgateway([a-z-]*)?.io|slsgateway.com)")
}

func extractPath(host, path string) string {
	if hostedDomainPattern.Copy().MatchString(host) {
		subdomain := strings.Split(host, ".")[0]
		return basePath + subdomain + path
	}
	return path
}

func systemPathFromSpace(space string) string {
	return basePath + space + "/"
}

// systemPathFromURL constructs system event path based on hostname and path
// on which the event was emitted. Helpful for "event.received" system event.
func systemPathFromURL(host, path string) string {
	if hostedDomainPattern.Copy().MatchString(host) {
		segment := strings.Split(path, "/")[1]
		return basePath + segment + "/"
	}
	return basePath
}
