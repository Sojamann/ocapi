package image

import (
	"strings"
)

// All images MUST specified as registry.com/namespace/image:tag
func parseParts(name string) (string, string, string) {
	// TODO: dont ignore anything here!!
	registry, rest, _ := strings.Cut(name, "/")
	image, tag, _ := strings.Cut(rest, ":")

	return registry, image, tag
}
