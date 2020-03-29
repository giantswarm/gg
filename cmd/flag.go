package cmd

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	fields []string
	greps  []string
	output string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&f.fields, "field", "f", nil, "Fields the output lines should contain only.")
	cmd.PersistentFlags().StringSliceVarP(&f.greps, "grep", "g", nil, "Grep for lines based on the given key:val regular expression.")
	cmd.PersistentFlags().StringVarP(&f.output, "output", "o", "json", "Output format, either json or text.")
}

func (f *flag) Validate() error {
	// Validate -f/--fields flags.
	for _, f := range f.fields {
		if f == "" {
			return microerror.Maskf(invalidFlagsError, "-f/--field must not be empty")
		}
		_, err := regexp.Compile(f)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Validate -g/--grep flags.
	for _, g := range f.greps {
		split := strings.Split(g, ":")

		if len(split) != 2 {
			return microerror.Maskf(invalidFlagsError, "%#q must have format key:val", g)
		}

		for _, s := range split {
			_, err := regexp.Compile(s)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	// Validate -o/--output flag.
	{
		if f.output == "" {
			return microerror.Maskf(invalidFlagsError, "-o/--output must not be empty")
		}
		if f.output != "json" && f.output != "text" {
			return microerror.Maskf(invalidFlagsError, "-o/--output must either be text or json")
		}
	}

	return nil
}
