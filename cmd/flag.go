package cmd

import (
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
)

type flag struct {
	grep string
}

func (f *flag) Init(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&f.grep, "grep", "g", "", "Grep for lines based on the given regular expression.")
}

func (f *flag) Validate() error {
	_, err := regexp.Compile(f.grep)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
