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
	extracted := path
	if hostedDomainPattern.Copy().MatchString(host) {
		subdomain := strings.Split(host, ".")[0]
		extracted = basePath + subdomain + path
	}
	return extracted
}

func systemEventPath(space string) string {
	return basePath + space + basePath
}
