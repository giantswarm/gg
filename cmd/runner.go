package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

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

	var expressions [][]*regexp.Regexp
	{
		for _, g := range r.flag.greps {
			split := strings.Split(g, ":")

			var pair []*regexp.Regexp
			pair = append(pair, regexp.MustCompile(split[0]))
			pair = append(pair, regexp.MustCompile(split[1]))
			expressions = append(expressions, pair)
		}
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
			match, err := matcher.Match(l, expressions)
			if err != nil {
				return microerror.Mask(err)
			}

			if !match {
				continue
			}
		}

		fmt.Fprint(r.stdout, l)
	}

	err = scanner.Err()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
