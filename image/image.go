package image

import (
	"fmt"
	"strings"

	"github.com/sojamann/opcapi/registry"
)

type Image struct {
	registryHost string
	name         string
	tag          string
	architecture string
	layers       []string
}

func ImageFromManifest(registryHost string, mp *registry.Manifest) *Image {
	layers := make([]string, 0, len(mp.FsLayers))

	for _, layer := range mp.FsLayers {
		layers = append(layers, layer.BlobSum)
	}
	return &Image{
		registryHost: registryHost,
		name:         mp.Name,
		tag:          mp.Tag,
		architecture: mp.Architecture,
		layers:       layers,
	}
}

func (image *Image) FullyQualifiedName() string {
	return fmt.Sprintf("%s/%s:%s", image.registryHost, image.name, image.tag)
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
	width := len(image.layers[0])
	center := func(s string) string {
		return fmt.Sprintf("%*s", -width, fmt.Sprintf("%*s", (width+len(s))/2, s))
	}

	builder := strings.Builder{}

	seperator := strings.Repeat("-", width+4) + "\n"

	builder.WriteString(seperator)
	builder.WriteString("| ")
	builder.WriteString(center(fmt.Sprintf("[ %s ]", image.name+":"+image.tag)))
	builder.WriteString(" |")
	builder.WriteRune('\n')
	builder.WriteString(seperator)

	for _, layer := range image.layers {
		builder.WriteString("| ")
		builder.WriteString(center(layer))
		builder.WriteString(" |")
		builder.WriteRune('\n')
	}
	builder.WriteString(seperator)
	return builder.String()
}
