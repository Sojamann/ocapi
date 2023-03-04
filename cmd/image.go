package cmd

import (
	"fmt"
	"os"

	"github.com/sojamann/opcapi/image"
	"github.com/sojamann/opcapi/registry"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "image short desc",
	Long:  "image long desc",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(flagDockerConfig)
	},
}

var imageLsCmd = &cobra.Command{
	Use:   "ls pattern",
	Short: "ls short desc",
	Long:  "ls long desc",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imagePattern := image.ImagePattern(args[0])
		if !imagePattern.IsValid() {
			fmt.Fprintln(os.Stderr, "Image pattern is not valid")
			os.Exit(1)
		}

		registryHost := imagePattern.RegistryHost()

		r, err := registry.NewRegisty(registryHost)

		_, imageSpec, tagSpec := image.ParseParts(string(imagePattern))
		specifiers, err := image.ExpandGlob(r, imageSpec, tagSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Err: %v\n", err)
			os.Exit(1)
		}

		for _, s := range specifiers {
			fmt.Printf("%s/%s:%s\n", registryHost, s.ImageName, s.Tag)
		}
	},
}

var imageShowCmd = &cobra.Command{
	Use:   "show registry/image:tag",
	Short: "show short desc",
	Long:  "show long desc",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryHost, imageName, tag := image.ParseParts(args[0])

		r, err := registry.NewRegisty(registryHost)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Err: %v\n", err)
			os.Exit(1)
		}

		manifest, err := r.GetManifest(imageName, tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Err: %v\n", err)
			os.Exit(1)
		}

		img := image.ImageFromManifest(manifest)
		fmt.Println(img)
	},
}

func init() {
	imageCmd.AddCommand(imageLsCmd)
	imageCmd.AddCommand(imageShowCmd)

	rootCmd.AddCommand(imageCmd)
}
