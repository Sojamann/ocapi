package image

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sojamann/opcapi/registry"
)

type ImageSpecifier struct {
	Registry  *registry.Registry
	ImageName string
	Tag       string
}

type ImagePattern string

func (s *ImagePattern) IsValid() bool {
	registryPattern := `[\w.-_]+`
	imagePattern := `(\*|(([\w-_]+)+)(\*|\/)?)`
	tagPattern := `(\*|[\w-_]+)`
	r := regexp.MustCompile(fmt.Sprintf("%s/%s:%s", registryPattern, imagePattern, tagPattern))
	return r.MatchString(string(*s))
}

// TODO: there comments are not correct really anymore
// It is a glob if the tag is given as asterix (*)
// It is a glob if the name ends with /*
// It is a glob if no tag is given
// TODO: this is sequential as of now
func (s *ImagePattern) Expand() ([]ImageSpecifier, error) {
	registryHost, imageSpecifier, tagSpecifier := ParseParts(string(*s))

	r, err := registry.NewRegisty(registryHost)
	if err != nil {
		return nil, err
	}

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
			result = append(result, ImageSpecifier{r, image, tagSpecifier})
		}
		tags, err := r.GetTags(image)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags {
			result = append(result, ImageSpecifier{r, image, tag})
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
