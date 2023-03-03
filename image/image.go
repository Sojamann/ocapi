package image

import (
	"strings"

	"github.com/sojamann/opcapi/registry"
)

type Image struct {
	name         string
	tag          string
	architecture string
	layers       []string
}

func ImageFromManifest(mp *registry.Manifest) *Image {
	layers := make([]string, 0, len(mp.FsLayers))

	for _, layer := range mp.FsLayers {
		layers = append(layers, layer.BlobSum)
	}

	return &Image{
		name:         mp.Name,
		tag:          mp.Tag,
		architecture: mp.Architecture,
		layers:       layers,
	}
}

// For this function to return true parent must be a true base image
// parent = [a, b, c, d]
// child  = [a, b, c, d, e, f]
func (image *Image) IsParentOf(child *Image) bool {
	if len(image.layers) > len(child.layers) {
		return false
	}

	offset := len(child.layers) - len(image.layers)

	for i := len(image.layers) - 1; i >= 0; i-- {
		// add offset as child has more layers
		cLayer := child.layers[i+offset]
		pLayer := image.layers[i]
		if pLayer != cLayer {
			return false
		}
	}

	return true
}

func (image *Image) String() string {
	builder := strings.Builder{}
	builder.WriteString(image.name + ":" + image.tag)
	builder.WriteRune('\n')
	for _, layer := range image.layers {
		builder.WriteString(layer)
		builder.WriteRune('\n')
	}
	return builder.String()
}
