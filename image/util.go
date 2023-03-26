package image

import (
	"os"
	"strings"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
)

// All images MUST specified as registry.com/namespace/image:tag
func parseParts(name string) (string, string, string) {
	// TODO: dont ignore anything here!!
	registry, rest, _ := strings.Cut(name, "/")
	image, tag, _ := strings.Cut(rest, ":")

	return registry, image, tag
}

// Returns a new progressbar (a slightly modified progressbar.Default)
func pbar(desc string, n int) *progressbar.ProgressBar {
	return progressbar.NewOptions(
		n,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}
