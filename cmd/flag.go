package cmd

import (
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	greps []string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&f.greps, "grep", "g", nil, "Grep for lines based on the given key:val regular expression.")
}

func (f *flag) Validate() error {
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

	return nil
}
