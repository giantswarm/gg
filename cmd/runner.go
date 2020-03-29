package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gg/pkg/formatter"
	"github.com/giantswarm/gg/pkg/matcher"
	"github.com/giantswarm/gg/pkg/splitter"
)

type runner struct {
	flag   *flag
	stdin  io.Reader
	stdout io.Writer
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

	scanner := bufio.NewScanner(r.stdin)
	scanner.Split(splitter.New().Split)

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
			match, err := matcher.Match(l, r.flag.greps)
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

		// Transform the current line of the stream based on the given fields with
		// the -f/--fields flag. We only want to print lines containing given
		// fields.
		if len(r.flag.fields) != 0 {
			newLine, err := formatter.Fields(l, r.flag.fields)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		{
			newLine, err := formatter.Output(l, r.flag.output)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		// Transform the current line of the stream so that it is colourized.
		{
			newLine, err := formatter.Colour(l, r.flag.output)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		fmt.Fprint(r.stdout, l)
	}

	err = scanner.Err()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
