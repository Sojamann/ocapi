package image

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sojamann/ocapi/registry"
	"github.com/sojamann/ocapi/sliceops"
)

const registryPattern = `[\w.-_]+`
const imagePattern = `(\*|([\w-_./]+)(\*|\/)?)`
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

// expands an image specifier to a list of images
// all images: *
// all images below this path: some/image/*
// all images below this path: some/image/
// full image path no need to expand: some/image/image
func expandImageSpecifier(r *registry.Registry, imageSpecifier string) ([]string, error) {
	images := make([]string, 0, 1)

	// if there is no glob syntax -> we can end it here
	if !strings.HasSuffix(imageSpecifier, "*") && !strings.HasSuffix(imageSpecifier, "/") {
		images = append(images, imageSpecifier)
		return images, nil
	}

	images, err := r.GetCatalog()
	if err != nil {
		return nil, err
	}

	matcher := regexp.MustCompile(strings.TrimRight(imageSpecifier, "*") + ".*")
	for _, image := range images {
		if matcher.MatchString(image) {
			images = append(images, image)
		}
	}
	return images, nil
}

func expandTagSpecifier(r *registry.Registry, images []string, tagSpecifier string) ([]ImageSpecifier, error) {
	imageSpecifiers := make([]ImageSpecifier, 0, 1)

	// when the tag is specified add the tag to all images but make sure
	// that the tag exists for the image
	if tagSpecifier != "*" {
		for _, image := range images {
			is := ImageSpecifier{r, image, tagSpecifier}
			exists, err := is.Exists()
			if err != nil {
				return nil, err
			}
			if exists {
				imageSpecifiers = append(imageSpecifiers, is)
			}

		}
		return imageSpecifiers, nil
	}

	type tagFetchResult struct {
		image string
		tags  []string
		err   error
	}
	tagFetchResults := sliceops.MapAsync(images, func(image string) tagFetchResult {
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

// TODO: there comments are not correct really anymore
// It is a glob if the tag is given as asterix (*)
// It is a glob if the name ends with /*
// It is a glob if no tag is given
// TODO: this is sequential as of now
func (s *ImagePattern) ExpandToSpecifiers() ([]ImageSpecifier, error) {
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
	matchingImageNames, err := expandImageSpecifier(r, imageSpecifier)
	if err != nil {
		return nil, err
	}
	return expandTagSpecifier(r, matchingImageNames, tagSpecifier)
}

func (s *ImagePattern) ExpandToImages() ([]*Image, error) {
	specifiers, err := s.ExpandToSpecifiers()
	if err != nil {
		return nil, err
	}

	type result struct {
		img *Image
		err error
	}
	imageGetResult := sliceops.MapAsync(specifiers, func(sp ImageSpecifier) result {
		img, err := sp.ToImage()
		return result{img, err}
	})

	images := make([]*Image, 0, len(specifiers))
	for _, result := range imageGetResult {
		if result.err != nil {
			return nil, result.err
		}

		images = append(images, result.img)
	}

	return images, nil
}
