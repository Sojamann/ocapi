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

var imageBasedOnCmd = &cobra.Command{
	Use:   "based-on registry/image:tag registry/images/*:*",
	Short: "based-on short desc",
	Long:  "based-on long desc",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		childImgSpecifier, err := image.ImageSpecifierParse(args[0])
		if err != nil {
			return err
		}

		parentImagePattern := image.ImagePattern(args[1])
		if !parentImagePattern.IsValid() {
			return fmt.Errorf("image pattern '%s' does not seem to be valid", args[1])
		}

		parentImgSpecifiers, err := parentImagePattern.Expand()
		if err != nil {
			return err
		}

		img, err := childImgSpecifier.ToImage()
		if err != nil {
			return err
		}

		matched := false
		for _, parentImgSpecifier := range parentImgSpecifiers {
			parentImg, err := parentImgSpecifier.ToImage()
			if err != nil {
				return err
			}

			if parentImg.IsParentOf(img) {
				fmt.Println(parentImgSpecifier)
				matched = true
			}
		}

		if matched {
			os.Exit(0)
		} else {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageLsCmd)
	imageCmd.AddCommand(imageShowCmd)
	imageCmd.AddCommand(imageBasedOnCmd)

	rootCmd.AddCommand(imageCmd)
}
