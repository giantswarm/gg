package cmd

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	fields  []string
	group   string
	output  string
	selects []string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&f.fields, "field", "f", nil, "Fields the output lines should contain only.")
	cmd.PersistentFlags().StringVarP(&f.group, "group", "g", "", "Group logs by inserting an empty line after the group end.")
	cmd.PersistentFlags().StringVarP(&f.output, "output", "o", "json", "Output format, either json or text.")
	cmd.PersistentFlags().StringSliceVarP(&f.selects, "select", "s", nil, "Select lines based on the given key:val regular expression.")
}

func (f *flag) Validate() error {
	// Validate -f/--field flags.
	for _, f := range f.fields {
		if f == "" {
			return microerror.Maskf(invalidFlagsError, "-f/--field must not be empty")
		}
		if len(f) < 3 {
			return microerror.Maskf(invalidFlagsError, "-f/--field must at least be 3 characters long")
		}
		_, err := regexp.Compile(f)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Validate -g/--group flag.
	if f.group != "" {
		if len(f.group) < 3 {
			return microerror.Maskf(invalidFlagsError, "-g/--group must at least be 3 characters long")
		}
		_, err := regexp.Compile(f.group)
		if err != nil {
			return microerror.Mask(err)
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

	// Validate -s/--select flags.
	for _, s := range f.selects {
		split := strings.Split(s, ":")

		if len(split) != 2 {
			return microerror.Maskf(invalidFlagsError, "-s/--select must have format key:val")
		}

		for _, s := range split {
			if len(s) < 3 {
				return microerror.Maskf(invalidFlagsError, "-s/--select key-val must at least be 3 characters long respectively")
			}
			_, err := regexp.Compile(s)
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}
