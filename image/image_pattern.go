package image

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/sojamann/opcapi/registry"
)

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
	if !s.IsValid() {
		panic("Expanding not validated ImagePattern is not okay .....")
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

	result := make([]ImageSpecifier, 0, 1)

	// if tagSpecifier is not a wildcard end this function early
	if tagSpecifier != "*" {
		for _, image := range imagesToCheck {
			result = append(result, ImageSpecifier{r, image, tagSpecifier})
		}

		return result, nil
	}

	// fetch all tags of the images in parallel
	var wg sync.WaitGroup
	var lock sync.Mutex
	wg.Add(len(imagesToCheck))
	for _, image := range imagesToCheck {
		go func(image string) {
			defer wg.Done()

			tags, err := r.GetTags(image)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed getting tags for %s due to: %v", image, err)
				return
			}

			lock.Lock()
			for _, tag := range tags {
				result = append(result, ImageSpecifier{r, image, tag})
			}
			lock.Unlock()
		}(image)
	}

	wg.Wait()

	return result, nil
}
