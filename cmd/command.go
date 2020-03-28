package cmd

import (
	"os"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gg/cmd/version"
	"github.com/giantswarm/gg/pkg/project"
)

type Config struct {
}

func New(config Config) (*cobra.Command, error) {
	var err error

	var versionCmd *cobra.Command
	{
		c := version.Config{}

		versionCmd, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	f := &flag{}

	r := &runner{
		flag:   f,
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}

	c := &cobra.Command{
		Use:          project.Name(),
		Short:        project.Description(),
		Long:         project.Description(),
		RunE:         r.Run,
		SilenceUsage: true,
	}

	f.Init(c)

	c.AddCommand(versionCmd)

	return c, nil
}
