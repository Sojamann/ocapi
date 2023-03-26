package image

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/life4/genesis/slices"
	"github.com/rs/zerolog/log"
	"github.com/sojamann/ocapi/registry"
)

const numConcurrentTasks = 5

const registryPattern = `[\w.-_]+`
const imagePattern = `(\*|([\w-_./]+)(\*|\/)?)`
const tagPattern = `(\*|[\w-_.]+)`

var imagePatternRe = regexp.MustCompile(fmt.Sprintf("^%s/%s:%s$", registryPattern, imagePattern, tagPattern))

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

type ImagePattern string

func (s *ImagePattern) IsValid() bool {
	return ValidateImagePattern(string(*s)) == nil
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

	log.Debug().Str("pattern", string(*s)).Msg("expanding image pattern")
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

	bar := pbar("Getting image tags", len(matchingImageNames))
	defer bar.Clear()
	type result struct {
		is  []ImageSpecifier
		err error
	}
	tagResults := slices.MapAsync(matchingImageNames, numConcurrentTasks, func(image string) result {
		is, err := expandTagSpecifier(r, image, tagSpecifier)
		bar.Add(1)
		return result{is, err}
	})

	imageSpecifiers := make([]ImageSpecifier, 0, len(matchingImageNames))
	for _, result := range tagResults {
		if result.err != nil {
			return nil, err
		}
		imageSpecifiers = append(imageSpecifiers, result.is...)
	}

	return imageSpecifiers, nil
}

func (s *ImagePattern) ExpandToImages() ([]*Image, error) {
	specifiers, err := s.ExpandToSpecifiers()
	if err != nil {
		return nil, err
	}

	bar := pbar("Getting image tags", len(specifiers))
	defer bar.Clear()
	type result struct {
		img *Image
		err error
	}
	imageGetResult := slices.MapAsync(specifiers, numConcurrentTasks, func(sp ImageSpecifier) result {
		img, err := sp.ToImage()
		bar.Add(1)
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

// expands an image name to a list of images
// all images: *
// all images below this path: some/image/*
// all images below this path: some/image/
// full image path no need to expand: some/image/image
func expandImageSpecifier(r *registry.Registry, imageName string) ([]string, error) {
	log.Debug().Str("host", r.Host).Str("image", imageName).Msg("expanding image name")

	images := make([]string, 0, 1)

	// if there is no glob syntax -> we can end it here
	if !strings.HasSuffix(imageName, "*") && !strings.HasSuffix(imageName, "/") {
		images = append(images, imageName)
		return images, nil
	}

	images, err := r.GetCatalog()
	if err != nil {
		return nil, err
	}

	matcher := regexp.MustCompile("^" + strings.TrimRight(imageName, "*") + ".*")
	return slices.Filter(images, matcher.MatchString), nil
}

func expandTagSpecifier(r *registry.Registry, image string, tagName string) ([]ImageSpecifier, error) {
	log.Debug().Str("host", r.Host).Str("image", image).Str("tag", tagName).Msg("expanding tag")
	imageSpecifiers := make([]ImageSpecifier, 0, 1)

	// when the tag is specified add the tag to all images but make sure
	// that the tag exists for the image
	if tagName != "*" {
		is := ImageSpecifier{r, image, tagName}
		exists, err := is.Exists()
		if err != nil {
			return nil, err
		}
		if exists {
			imageSpecifiers = append(imageSpecifiers, is)
		}

		return imageSpecifiers, nil
	}

	tags, err := r.GetTags(image)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		imageSpecifiers = append(imageSpecifiers, ImageSpecifier{r, image, tag})
	}

	return imageSpecifiers, nil
}
