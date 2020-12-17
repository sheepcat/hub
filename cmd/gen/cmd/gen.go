package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tinkerbell/actions/pkg/artifacthub"
)

type generateOptions struct {
	context string
	output  string
}

var generateOpts = &generateOptions{}

var generateCmd = &cobra.Command{
	Use:   "generate [--context .]",
	Short: "Generate the static webiste",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(generateOpts)
	},
}

func init() {
	generateCmd.PersistentFlags().StringVar(&generateOpts.context, "context", ".", "base path for the proposals repository in your local file system")
	generateCmd.PersistentFlags().StringVar(&generateOpts.output, "output", "./artifacthub-manifests", "where the generate website will be stored")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(opts *generateOptions) error {
	actionsDir := path.Join(opts.context, "actions")
	info, err := os.Stat(actionsDir)
	if os.IsNotExist(err) {
		return errors.Wrap(err, "we expect a actions directory inside the repository.")
	}
	if info.IsDir() == false {
		return errors.New("the expected actions directory has to be a directory, not a file")
	}

	files, err := ioutil.ReadDir(actionsDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(generateOpts.output); os.IsNotExist(err) {
		os.Mkdir(generateOpts.output, 0700)
	}

	// This is the manifest with all the pre-populated information. Ideally
	// this has to be populated from an outside configuration file.
	manifest := &artifacthub.Manifest{
		Provider: struct {
			Name string `yaml:"name"`
		}{
			Name: "tinkerbell-community",
		},
		HomeURL:  "https://github.com/tinkerbell/actions",
		LogoPath: "./../../logo.png",
		License:  "Apache-2",
		Links: []struct {
			Name string `yaml:"name"`
			URL  string `yaml:"url"`
		}{
			{
				Name: "website",
				URL:  "https://tinkerbell.org/",
			},
			{
				Name: "support",
				URL:  "https://github.com/tinkerbell/actions/issues",
			},
		},
	}

	for _, f := range files {
		readmeFile, err := os.Open(path.Join(actionsDir, f.Name(), "README.md"))
		if err != nil {
			return errors.Wrap(err, "error reading the README.md proposal")
		}

		if err := artifacthub.PopulateFromActionMarkdown(readmeFile, manifest); err != nil {
			return errors.Wrap(err, "error converting the README.md to an ArtifactHub manifest")
		}

		if err := artifacthub.WriteToFile(manifest, generateOpts.output); err != nil {
			return errors.Wrap(err, "error writing manifest to a file")
		}
	}

	return nil
}
