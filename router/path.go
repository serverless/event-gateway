// +build !hosted

package router

func extractPath(host, path string) string {
	return path
}

func systemPathFromSpace(space string) string {
	return basePath
}

func systemPathFromPath(path string) string {
	return basePath
}
