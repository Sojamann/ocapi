package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// All images MUST specified as registry.com/namespace/image:tag
func parseParts(name string) (string, string, string) {
	// TODO: dont ignore anything here!!
	registry, rest, _ := strings.Cut(name, "/")
	image, tag, _ := strings.Cut(rest, ":")

	return registry, image, tag
}

type ImageSpecifier struct {
	ImageName string
	Tag       string
}

// TODO: there comments are not correct really anymore
// It is a glob if the tag is given as asterix (*)
// It is a glob if the name ends with /*
// It is a glob if no tag is given
func expandGlob(r *Registry, imageSpecifier, tagSpecifier string) ([]ImageSpecifier, error) {

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

func main() {
	credMap, err := LoadCredentialsFromDockerConfig("~/.docker/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	imageGlob := ""
	host, image, tag := parseParts(imageGlob)

	credentials, found := credMap[host]
	if !found {
		log.Fatalf("No creds for registry %s", host)
	}

	r, err := NewRegisty(host, credentials)
	if err != nil {
		log.Fatalln(err)
	}

	specifiers, err := expandGlob(r, image, tag)
	for _, specifier := range specifiers {
		fmt.Println(specifier)
	}

}
