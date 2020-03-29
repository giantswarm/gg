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
	var group string

	scanner := bufio.NewScanner(r.stdin)
	scanner.Split(splitter.New().Split)

	for scanner.Scan() {
		l := scanner.Text()

		// Filter the current line of the stream based on the given expression with
		// the -s/--select flag. We only want to print matching lines.
		{
			match, err := matcher.Match(l, r.flag.selects)
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

		// Transform the the log stream based on the given fields with the
		// -g/--group flag. We want to group lines according to their semnatical
		// meaning. For instance you want to group logs per resource implementation
		// or reconciliation loop. The grouping is simply done by inserting an empty
		// line.
		if r.flag.group != "" {
			value, err := matcher.Value(l, r.flag.group)
			if err != nil {
				return microerror.Mask(err)
			}

			// Initialize the group value with the first matching line.
			if group == "" {
				group = value
			}

			// As soon as we find a new group value we insert an empty line and
			// remember the new group value.
			if value != group {
				fmt.Fprint(r.stdout, "\n")
				group = value
			}
		}

		// Transform the current line of the stream based on the given fields with
		// the -f/--field flag. We only want to print lines containing given
		// fields.
		if len(r.flag.fields) != 0 {
			newLine, err := formatter.Fields(l, r.flag.fields)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		// Transform the current line of the stream based on the given fields with
		// the -o/--output flag. We only want to print selected fields.
		{
			newLine, err := formatter.Output(l, r.flag.output)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		// Transform the current line of the stream so that it is colourized.
		//
		// Note that certain control characters are inserted into the strings in
		// order to make them colorful. This implies that the JSON strings do not
		// contain valid JSON objects anymore. Therefore all JSON object related
		// operations must have been done at this point.
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
