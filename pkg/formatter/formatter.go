package formatter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"regexp"

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

//
//     if kv.Key == "time" {
//       t, err := time.Parse(timeFormatFrom, kv.Value.(string))
//       if err != nil {
//         return "", microerror.Mask(err)
//       }
//       fm.Set(kv.Key, t.Format(timeFormatTo))
//     }

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
			_, err := io.WriteString(b, l)
			if err != nil {
				return "", microerror.Mask(err)
			}
		default:
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): (".*")(,?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+": "+p.Value("$3")+"$4")
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): (.*)(,?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+": "+p.Value("$3")+"$4")
			l = regexp.MustCompile(`(?m)^([\s]*)("[\w-.]*"): ([\{\[]?)$`).ReplaceAllString(l, "$1"+p.Key("$2")+" $3")

			_, err := io.WriteString(b, l)
			if err != nil {
				return "", microerror.Mask(err)
			}
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

	f := fm.EntriesIter()
	for {
		kv, ok := f()
		if !ok {
			break
		}

		if !containsExp(fields, kv.Key) {
			fm.Delete(kv.Key)
		}
	}

	newFormat, err := fm.MarshalJSON()
	if err != nil {
		return "", microerror.Mask(err)
	}

	// The line that comes in contains a newline at the end. When we transform it
	// the Feature Map does not maintain it. So before returning it we have to add
	// the newline back.
	return string(newFormat) + "\n", nil
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

func containsExp(fields []string, field string) bool {
	for _, f := range fields {
		if regexp.MustCompile(f).MatchString(field) {
			return true
		}
	}

	return false
}
