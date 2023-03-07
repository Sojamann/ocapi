package image

import (
	"fmt"
	"regexp"

	"github.com/sojamann/ocapi/registry"
)

type ImageSpecifier struct {
	Registry  *registry.Registry
	ImageName string
	Tag       string
}

var imageSpecifierRe = regexp.MustCompile(`^[\w.-_]+\/([\w-_.]+\/)*[\w-_.]+:[\w-_.]+$`)

type InvalidImageSpecifier string

func (s InvalidImageSpecifier) Error() string {
	return fmt.Sprintf("'%s' is not a valid image specifier (%s)", string(s), imageSpecifierRe.String())
}

func ValidateImageSpecifier(s string) error {
	if imageSpecifierRe.MatchString(s) {
		return nil
	}
	return InvalidImageSpecifier(s)
}

func ImageSpecifierParse(s string) (*ImageSpecifier, error) {
	if !imageSpecifierRe.MatchString(s) {
		return nil, InvalidImageSpecifier(s)
	}

	registryHost, imageName, tag := parseParts(s)

	r, err := registry.NewRegisty(registryHost)
	if err != nil {
		return nil, err
	}

	return &ImageSpecifier{
		Registry:  r,
		ImageName: imageName,
		Tag:       tag,
	}, nil
}

func (is *ImageSpecifier) Exists() (bool, error) {
	return is.Registry.Exists(is.ImageName, is.Tag)
}

func (is *ImageSpecifier) ToImage() (*Image, error) {
	manifest, err := is.Registry.GetManifest(is.ImageName, is.Tag)
	if err != nil {
		return nil, err
	}

	return ImageFromManifest(is.Registry.Host, manifest), nil
}

func (is ImageSpecifier) String() string {
	return fmt.Sprintf("%s/%s:%s", is.Registry.Host, is.ImageName, is.Tag)
}
