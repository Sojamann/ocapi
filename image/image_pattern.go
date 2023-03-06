package image

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sojamann/opcapi/registry"
	"github.com/sojamann/opcapi/sliceops"
)

const registryPattern = `[\w.-_]+`
const imagePattern = `(\*|(([\w-_]+)+)(\*|\/)?)`
const tagPattern = `(\*|[\w-_.]+)`

var imagePatternRe = regexp.MustCompile(fmt.Sprintf("^%s/%s:%s$", registryPattern, imagePattern, tagPattern))

type ImagePattern string

type InvalidImagePattern string

func (s InvalidImagePattern) Error() string {
	return fmt.Sprintf("'%s' is not a valid image pattern (%s)", string(s), imagePatternRe.String())
}

func ValidateImagePattern(s string) error {
	if imagePatternRe.MatchString(s) {
		return nil
	}
	return InvalidImagePattern(s)
}

func (s *ImagePattern) IsValid() bool {
	return ValidateImagePattern(string(*s)) == nil
}

// TODO: there comments are not correct really anymore
// It is a glob if the tag is given as asterix (*)
// It is a glob if the name ends with /*
// It is a glob if no tag is given
// TODO: this is sequential as of now
func (s *ImagePattern) Expand() ([]ImageSpecifier, error) {
	// TODO: maybe return an error instead of a panic??
	if !s.IsValid() {
		panic("Expanding non validated ImagePattern is not okay .....")
	}

	registryHost, imageSpecifier, tagSpecifier := parseParts(string(*s))

	r, err := registry.NewRegisty(registryHost)
	if err != nil {
		return nil, err
	}

	// resolve imageSpecifier
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

	imageSpecifiers := make([]ImageSpecifier, 0, 1)

	// if tagSpecifier is not a wildcard end this function early
	if tagSpecifier != "*" {
		for _, image := range imagesToCheck {
			imageSpecifiers = append(imageSpecifiers, ImageSpecifier{r, image, tagSpecifier})
		}

		return imageSpecifiers, nil
	}

	type tagFetchResult struct {
		image string
		tags  []string
		err   error
	}
	tagFetchResults := sliceops.MapAsync(imagesToCheck, func(image string) tagFetchResult {
		tags, err := r.GetTags(image)
		return tagFetchResult{image, tags, err}
	})

	for _, res := range tagFetchResults {
		if res.err != nil {
			return nil, res.err
		}
		for _, tag := range res.tags {
			imageSpecifiers = append(imageSpecifiers, ImageSpecifier{r, res.image, tag})
		}
	}

	return imageSpecifiers, nil
}
