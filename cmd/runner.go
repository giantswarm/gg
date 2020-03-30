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
	var isErr bool
	var hasNewLine bool

	scanner := bufio.NewScanner(r.stdin)
	scanner.Split(splitter.New().Split)

	for scanner.Scan() {
		l := scanner.Text()

		// Check if the current line of the stream has the expected format of our
		// JSON log objects. If it does not appear to be valid JSON, we simply print
		// the line as it is.
		//
		// Note that for invalid JSON messages we print an extra line before and
		// after the printed text. In this case we remember that an empty line got
		// already printed, so that further grouping of logs doesn't introduce any
		// unnecessary padding.
		{
			if l[0] != '{' {
				if !hasNewLine {
					fmt.Fprint(r.stdout, "\n")
				}
				fmt.Fprint(r.stdout, l)
				fmt.Fprint(r.stdout, "\n")

				hasNewLine = true

				continue
			}
		}

		// We want to print error logs differently. Therefore we check if the
		// current line of the stream is what we expect to be an error log early on
		// in the processing so that relevant information for detection are not
		// removed before we get the chance to inspect the complete line.
		{
			isErr, err = formatter.IsErr(l)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -f/--field flag. We do not want to print lines that do not have the
		// fields we want to display.
		//
		// TODO we check !isErr when checking for a match which is because of legacy
		// microerror structures where the annotation is magically reverse
		// engineered from the legacy stack. Once we do not have to deal with these
		// legacy structures we can remove the !isErr check.
		{
			match, err := matcher.Match(l, matcher.Exp(r.flag.fields))
			if err != nil {
				return microerror.Mask(err)
			}

			if !isErr && !match {
				continue
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -g/--group flag. We do not want to print lines that do not have the
		// fields we want to group by.
		{
			match, err := matcher.Match(l, matcher.Exp([]string{r.flag.group}))
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

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
			//
			// Note that a new line is only inserted in case no invalid JSON got
			// detected. This is to prevent unnecessary extra padding.
			if value != group {
				if !hasNewLine {
					fmt.Fprint(r.stdout, "\n")
				}
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
		if isErr {
			newLine, err := formatter.Error(l, r.flag.output, r.flag.fields)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		} else {
			newLine, err := formatter.Colour(l, r.flag.output)
			if err != nil {
				return microerror.Mask(err)
			}

			l = newLine
		}

		// Finally we print the current line of the stream based on its processed
		// selection and transformation.
		//
		// Note that we reset hasNewLine again to start all over with the detection
		// of extra padding. This is basically for eye candy reasons.
		{
			hasNewLine = false
			fmt.Fprint(r.stdout, l)
		}
	}

	err = scanner.Err()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
