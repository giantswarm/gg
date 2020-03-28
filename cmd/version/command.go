package version

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

const (
	name        = "version"
	description = "Print version information."
)

type Config struct {
	Stdout io.Writer
}

func New(config Config) (*cobra.Command, error) {
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}

	f := &flag{}

	r := &runner{
		flag:   f,
		stdout: config.Stdout,
	}

	c := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		RunE:  r.Run,
	}

	f.Init(c)

	return c, nil
}
