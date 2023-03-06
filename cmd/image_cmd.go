package cmd

import (
	"fmt"
	"os"

	"github.com/sojamann/opcapi/image"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Commands centered around ONE image",
	Long:  "Commands centered around ONE image",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(flagDockerConfig)
	},
}

var imageLsCmd = &cobra.Command{
	Use:   "ls pattern",
	Short: "List all images matching the pattern",
	Long:  "List all images matching the pattern",
	Args: cobra.MatchAll(
		cobra.ExactArgs(1),
		validateArgNo(0, image.ValidateImagePattern),
	),
	Run: func(cmd *cobra.Command, args []string) {
		imagePattern := image.ImagePattern(args[0])
		if !imagePattern.IsValid() {
			fmt.Fprintln(os.Stderr, "Image pattern is not valid")
			os.Exit(1)
		}

		specifiers, err := imagePattern.ExpandToSpecifiers()
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
	Short: "Show OCI image",
	Long:  "Show OCI image",
	Args: cobra.MatchAll(
		cobra.ExactArgs(1),
		validateArgNo(0, image.ValidateImageSpecifier),
	),
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
	Short: "Check parent images",
	Long:  "List all images matching the pattern on which the specified image a is based on",
	Args: cobra.MatchAll(
		cobra.ExactArgs(2),
		validateArgNo(0, image.ValidateImageSpecifier),
		validateArgNo(1, image.ValidateImagePattern),
	),
	RunE: func(cmd *cobra.Command, args []string) error {
		childImgSpecifier, err := image.ImageSpecifierParse(args[0])
		if err != nil {
			return err
		}
		childImg, err := childImgSpecifier.ToImage()
		if err != nil {
			return err
		}

		parentImgPattern := image.ImagePattern(args[1])
		parentImgs, err := parentImgPattern.ExpandToImages()
		if err != nil {
			return err
		}

		matched := false
		for _, parentImg := range parentImgs {
			if parentImg.IsParentOf(childImg) {
				fmt.Println(parentImg.FullyQualifiedName())
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

var imageBaseOfCmd = &cobra.Command{
	Use:   "base-of registry/image:tag registry/images/*:*",
	Short: "Check child images",
	Long:  "List all images matching the pattern on which the specified image a is the base of",
	Args: cobra.MatchAll(
		cobra.ExactArgs(2),
		validateArgNo(0, image.ValidateImageSpecifier),
		validateArgNo(1, image.ValidateImagePattern),
	),
	RunE: func(cmd *cobra.Command, args []string) error {
		parentImgSpecifier, err := image.ImageSpecifierParse(args[0])
		if err != nil {
			return err
		}
		parentImg, err := parentImgSpecifier.ToImage()
		if err != nil {
			return err
		}

		childImgPattern := image.ImagePattern(args[1])
		childImgs, err := childImgPattern.ExpandToImages()
		if err != nil {
			return err
		}

		matched := false
		for _, childImg := range childImgs {
			if parentImg.IsParentOf(childImg) {
				fmt.Println(childImg.FullyQualifiedName())
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
	imageCmd.AddCommand(imageBaseOfCmd)

	rootCmd.AddCommand(imageCmd)
}
