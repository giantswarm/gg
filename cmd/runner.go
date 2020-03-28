package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gg/pkg/formatter"
	"github.com/giantswarm/gg/pkg/splitter"
)

type runner struct {
	flag *flag
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(splitter.New().Split)

	var e *regexp.Regexp
	{
		e = regexp.MustCompile(r.flag.grep)
	}

	for scanner.Scan() {
		l := scanner.Text()

		// Get the current line of the stream based on its anticipated format for
		// further processing.
		{
			f := formatter.Is(l)
			if f == formatter.JSON || f == formatter.Text {
				l = scanner.Text()
			}
			if f == formatter.None {
				l += "\n"
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -g/--grep flag. We only want to print matching lines.
		{
			m := e.MatchString(l)
			if !m {
				continue
			}
		}

		fmt.Printf("%s", l)
	}

	err = scanner.Err()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
