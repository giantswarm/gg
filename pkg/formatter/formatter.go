package formatter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gg/pkg/colour"
	"github.com/giantswarm/gg/pkg/featuremap"
	"github.com/giantswarm/gg/pkg/splitter"
)

const (
	timeFormatFrom = "2006-01-02T15:04:05.999999-07:00"
	timeFormatTo   = "15:04:05"
)

const (
	indentNone = ""
	indentFour = "    "
)

func ColourJSON(l string, p colour.Palette) (string, error) {
	var scanner *bufio.Scanner
	{
		b := &bytes.Buffer{}
		err := json.Indent(b, []byte(l), "", indentFour)
		if err != nil {
			return "", microerror.Mask(err)
		}

		scanner = bufio.NewScanner(b)
		scanner.Split(splitter.New().Split)
	}

	b := &bytes.Buffer{}
	for scanner.Scan() {
		l := scanner.Text()
		switch {
		case l[0] == '{' || l[0] == '}' || l[0] == '[' || l[0] == ']':
			io.WriteString(b, l)
		default:
			l = regexp.MustCompile(`(".*"): `).ReplaceAllString(l, p.Key("$1")+": ")
			l = regexp.MustCompile(`: (".*")`).ReplaceAllString(l, ": "+p.Value("$1"))
			l = regexp.MustCompile(`: ([^"].*)`).ReplaceAllString(l, ": "+p.Value("$1"))
			io.WriteString(b, l)
		}
	}

	return b.String(), nil
}

func Fields(l string, fields []string) (string, error) {
	fm := featuremap.New()
	err := fm.UnmarshalJSON([]byte(l))
	if err != nil {
		return "", microerror.Mask(err)
	}

	var expressions []*regexp.Regexp
	for _, f := range fields {
		expressions = append(expressions, regexp.MustCompile(f))
	}

	for _, e := range expressions {
		f := fm.EntriesIter()
		for {
			kv, ok := f()
			if !ok {
				break
			}

			if e.MatchString(kv.Key) {
				if kv.Key == "time" {
					t, err := time.Parse(timeFormatFrom, kv.Value.(string))
					if err != nil {
						return "", microerror.Mask(err)
					}

					fm.Set(kv.Key, t.Format(timeFormatTo))
					continue
				}
			} else {
				fm.Delete(kv.Key)
			}
		}
	}

	newFormat, err := fm.MarshalJSON()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(newFormat), nil
}

func IsErr(l string) (bool, error) {
	fm := featuremap.New()
	err := fm.UnmarshalJSON([]byte(l))
	if err != nil {
		return false, microerror.Mask(err)
	}

	isErr := fm.Has("level") && fm.Get("level") == "error"
	isWar := fm.Has("level") && fm.Get("level") == "warning"

	return isErr || isWar, nil
}
