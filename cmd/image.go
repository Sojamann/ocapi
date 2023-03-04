package cmd

import (
	"fmt"
	"os"

	"github.com/sojamann/opcapi/image"
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

		specifiers, err := imagePattern.Expand()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Err: %v\n", err)
			os.Exit(1)
		}

		for _, s := range specifiers {
			fmt.Printf("%s/%s:%s\n", s.Registry.Host, s.ImageName, s.Tag)
		}
	},
}

var imageShowCmd = &cobra.Command{
	Use:   "show registry/image:tag",
	Short: "show short desc",
	Long:  "show long desc",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageSpecifier, err := image.ImageSpecifierParse(args[0])
		if err != nil {
			return err
		}

		img, err := imageSpecifier.ToImage()
		if err != nil {
			return err
		}

		fmt.Println(img)

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageLsCmd)
	imageCmd.AddCommand(imageShowCmd)

	rootCmd.AddCommand(imageCmd)
}
