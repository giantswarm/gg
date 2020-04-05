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
)

func Fields(fm *featuremap.FeatureMap, fields []string) (*featuremap.FeatureMap, error) {
	nfm := featuremap.New()

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
				nfm.Set(kv.Key, kv.Value)
			}
		}
	}

	return nfm, nil
}

func IndentWithColour(fm *featuremap.FeatureMap, p colour.Palette) (string, error) {
	var scanner *bufio.Scanner
	{
		byt, err := fm.MarshalJSON()
		if err != nil {
			return "", microerror.Mask(err)
		}

		buf := &bytes.Buffer{}
		err = json.Indent(buf, byt, "", "    ")
		if err != nil {
			return "", microerror.Mask(err)
		}

		scanner = bufio.NewScanner(buf)
		scanner.Split(splitter.New().Split)
	}

	b := &bytes.Buffer{}
	for scanner.Scan() {
		l := scanner.Text()
		switch {
		case l[0] == '{' || l[0] == '}' || l[0] == '[' || l[0] == ']':
			_, err := io.WriteString(b, l)
			if err != nil {
				return "", microerror.Mask(err)
			}
		default:
			// Match the start of objects and arrays.
			//
			//     "stack" {
			//     "stack" [
			//
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): ([\{\[]?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+" $3")

			// Match string key-value pairs.
			//
			//     "kind": "unknown",
			//     "resource": "basedomain"
			//
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): (".*")(,?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+": "+p.Value("$3")+"$4")

			// Match other key-value pairs.
			//
			//     "line": 217,
			//     "resources": null
			//
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): (.*)(,?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+": "+p.Value("$3")+"$4")

			_, err := io.WriteString(b, l)
			if err != nil {
				return "", microerror.Mask(err)
			}
		}
	}

	return b.String(), nil
}

func IsErr(fm *featuremap.FeatureMap) (bool, error) {
	isErr := fm.Has("level") && fm.Get("level") == "error"
	isWar := fm.Has("level") && fm.Get("level") == "warning"

	return isErr || isWar, nil
}

func Time(fm *featuremap.FeatureMap, timeFormat string) (*featuremap.FeatureMap, error) {
	f := fm.EntriesIter()
	for {
		kv, ok := f()
		if !ok {
			break
		}

		if kv.Key == "time" {
			t, err := time.Parse(timeFormatFrom, kv.Value.(string))
			if err != nil {
				return nil, microerror.Mask(err)
			}

			fm.Set(kv.Key, t.Format(timeFormat))
		}
	}

	return fm, nil
}
