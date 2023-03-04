package image

import (
	"fmt"
	"regexp"

	"github.com/sojamann/opcapi/registry"
)

type ImageSpecifier struct {
	Registry  *registry.Registry
	ImageName string
	Tag       string
}

var imageSpecifierRe = regexp.MustCompile(`[\w.-_]+\/([\w-_]+\/)*[\w-_]+:[\w-_]+`)

func ImageSpecifierParse(s string) (*ImageSpecifier, error) {
	if !imageSpecifierRe.MatchString(s) {
		return nil, fmt.Errorf("'%s' does not seem to be a valid image specifier", s)
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

func (is *ImageSpecifier) ToImage() (*Image, error) {
	manifest, err := is.Registry.GetManifest(is.ImageName, is.Tag)
	if err != nil {
		return nil, err
	}

	return ImageFromManifest(manifest), nil
}
