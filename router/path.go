// +build !hosted

package router

func extractPath(host, path string) string {
	return path
}

func systemEventPath(space string) string {
	return "/"
}
