package image

import (
	"github.com/sojamann/opcapi/registry"
)

type ImageSpecifier struct {
	Registry  *registry.Registry
	ImageName string
	Tag       string
}

func ImageSpecifierParse(s string) (*ImageSpecifier, error) {
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
