package image

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sojamann/opcapi/registry"
)

type ImagePattern string

func (s *ImagePattern) IsValid() bool {
	registryPattern := `[\w.-_]+`
	imagePattern := `(\*|(([\w-_]+)+)(\*|\/)?)`
	tagPattern := `(\*|[\w-_]+)`
	r := regexp.MustCompile(fmt.Sprintf("%s/%s:%s", registryPattern, imagePattern, tagPattern))
	return r.MatchString(string(*s))
}

func (s *ImagePattern) RegistryHost() string {
	registryHost, _, _ := ParseParts(string(*s))
	return registryHost
}

type ImageSpecifier struct {
	ImageName string
	Tag       string
}

// TODO: there comments are not correct really anymore
// It is a glob if the tag is given as asterix (*)
// It is a glob if the name ends with /*
// It is a glob if no tag is given
// TODO: this is sequential as of now
func ExpandGlob(r *registry.Registry, imageSpecifier, tagSpecifier string) ([]ImageSpecifier, error) {

	imagesToCheck := make([]string, 0, 1)
	if !strings.HasSuffix(imageSpecifier, "*") && !strings.HasSuffix(imageSpecifier, "/") {
		imagesToCheck = append(imagesToCheck, imageSpecifier)
	} else {
		images, err := r.GetCatalog()
		if err != nil {
			return nil, err
		}

		matcher := regexp.MustCompile(strings.TrimRight(imageSpecifier, "*") + ".*")
		for _, image := range images {
			if matcher.MatchString(image) {
				imagesToCheck = append(imagesToCheck, image)
			}
		}
	}

	result := make([]ImageSpecifier, 0, 1)
	for _, image := range imagesToCheck {
		if tagSpecifier != "*" {
			result = append(result, ImageSpecifier{image, tagSpecifier})
		}
		tags, err := r.GetTags(image)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags {
			result = append(result, ImageSpecifier{image, tag})
		}
	}

	return result, nil
}

// All images MUST specified as registry.com/namespace/image:tag
func ParseParts(name string) (string, string, string) {
	// TODO: dont ignore anything here!!
	registry, rest, _ := strings.Cut(name, "/")
	image, tag, _ := strings.Cut(rest, ":")

	return registry, image, tag
}
