package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/giantswarm/gg/pkg/colour"
	"github.com/giantswarm/gg/pkg/featuremap"
	"github.com/giantswarm/gg/pkg/formatter"
	"github.com/giantswarm/gg/pkg/matcher"
	"github.com/giantswarm/gg/pkg/splitter"
)

type runner struct {
	flag   *flag
	stdin  io.Reader
	stdout io.Writer
	viper  *viper.Viper
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

	scanner := bufio.NewScanner(r.stdin)
	scanner.Split(splitter.New().Split)

	for scanner.Scan() {
		// Check if the current line of the stream has the expected format of our
		// JSON log objects. If it does not appear to be valid JSON, we simply print
		// the line as it is.
		//
		// Note that for invalid JSON messages we print an extra line before and
		// after the printed text. In this case we remember that an empty line got
		// already printed, so that further grouping of logs doesn't introduce any
		// unnecessary padding.
		{
			l := scanner.Text()

			if l[0] != '{' {
				fmt.Fprint(r.stdout, l)
				continue
			}
		}

		var fm *featuremap.FeatureMap
		{
			fm = featuremap.New()

			err = fm.UnmarshalJSON(scanner.Bytes())
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// We want to print error logs differently. Therefore we check if the
		// current line of the stream is what we expect to be an error log early on
		// in the processing so that relevant information for detection are not
		// removed before we get the chance to inspect the complete line.
		{
			isErr, err = formatter.IsErr(fm)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -f/--field flag. We do not want to print lines that do not have the
		// fields we want to display.
		if len(r.flag.fields) != 0 {
			match, err := matcher.Match(fm, matcher.Exp(r.flag.fields))
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

		// Filter errors without stack and annotation fields, in case we are looking
		// for these fields. We do not want to print logs that do not have fields we
		// are actually looking for, even if they are errors.
		if len(r.flag.fields) != 0 {
			match, err := matcher.Match(fm, matcher.ExpWithout(r.flag.fields, "stack"))
			if err != nil {
				return microerror.Mask(err)
			}

			if !match && isErr {
				continue
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -g/--group flag. We do not want to print lines that do not have the
		// fields we want to group by.
		if r.flag.group != "" {
			match, err := matcher.Match(fm, matcher.Exp([]string{r.flag.group}))
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

		// Filter the current line of the stream based on the given expression with
		// the -s/--select flag. We only want to print matching lines.
		if len(r.flag.selects) != 0 {
			match, err := matcher.Match(fm, r.flag.selects)
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
			value, err := matcher.Value(fm, r.flag.group)
			if err != nil {
				return microerror.Mask(err)
			}

			// Initialize the group value with the first matching line.
			if group == "" {
				group = value
			}

			// As soon as we find a new group value we insert empty lines for visual
			// separation and remember the new group value.
			if value != group {
				fmt.Fprint(r.stdout, "\n")
				fmt.Fprint(r.stdout, "\n")
				fmt.Fprint(r.stdout, "\n")
				group = value
			}
		}

		// Transform the current line of the stream based on the given fields with
		// the -f/--field flag. We only want to print lines containing given
		// fields.
		if len(r.flag.fields) != 0 {
			fm, err = formatter.Fields(fm, r.flag.fields)
			if err != nil {
				return microerror.Mask(err)
			}

			if fm.Len() == 0 {
				continue
			}
		}

		// Replace the given timestamps with the given time format. This should make
		// it easier for humans to compare the times at which logs got emitted.
		if r.flag.time != "" {
			fm, err = formatter.Time(fm, r.flag.time)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Transform the current line of the stream so that it is colourized.
		//
		// Note that certain control characters are inserted into the strings in
		// order to make them colorful. This implies that the JSON strings do not
		// contain valid JSON objects anymore. Therefore all JSON object related
		// operations must have been done at this point.
		var l string
		if isErr {
			var p colour.Palette
			if r.flag.colour {
				p = colour.Palette{Key: colour.DarkRed, Value: colour.LightRed}
			} else {
				p = colour.NewNoColourPalette()
			}

			l, err = formatter.IndentWithColour(fm, p)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			var p colour.Palette
			if r.flag.colour {
				p = colour.Palette{Key: colour.DarkGreen, Value: colour.LightGreen}
			} else {
				p = colour.NewNoColourPalette()
			}

			l, err = formatter.IndentWithColour(fm, p)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Finally we print the current line of the stream based on its processed
		// selection and transformation.
		{
			fmt.Fprint(r.stdout, l+"\n")
		}
	}

	err = scanner.Err()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
