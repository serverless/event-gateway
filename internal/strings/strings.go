package strings

import "strings"

// EnsurePrefix ensures s starts with prefix.
func EnsurePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}

// EnsureSuffix ensures s ends with suffix.
func EnsureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}
