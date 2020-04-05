package cmd

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gg/pkg/config"
)

const (
	defaultColour = false
	defaultGroup  = ""
	defaultTime   = ""
)

type flag struct {
	colour  bool
	fields  []string
	group   string
	selects []string
	time    string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&f.colour, "colour", "c", config.Colour(defaultColour), "Whether to colourize printed output or not.")
	cmd.PersistentFlags().StringSliceVarP(&f.fields, "field", "f", nil, "Fields the output lines should contain only.")
	cmd.PersistentFlags().StringVarP(&f.group, "group", "g", config.Group(defaultGroup), "Group logs by inserting an empty line after the group end.")
	cmd.PersistentFlags().StringSliceVarP(&f.selects, "select", "s", nil, "Select lines based on the given key:val regular expression.")
	cmd.PersistentFlags().StringVarP(&f.time, "time", "t", config.Time(defaultTime), "Time format used to print timestamps.")
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

	// Validate -t/--time flag. Note that time can be empty, which means the
	// timestamp is not modified.

	return nil
}
