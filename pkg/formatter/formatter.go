package formatter

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gg/pkg/colour"
	"github.com/giantswarm/gg/pkg/featuremap"
)

const (
	timeFormatFrom = "2006-01-02T15:04:05.999999-07:00"
	timeFormatTo   = "15:04:05"
)

const (
	arrayEnd   = "]"
	arrayStart = "["
	mapEnd     = "}"
	mapStart   = "{"
)

const (
	indentNone = ""
	indentFour = "    "
)

var (
	colorKey    = color.New(color.FgHiBlue).SprintFunc()
	colorString = color.New(color.FgGreen).SprintFunc()
	colorError  = color.New(color.FgRed).SprintFunc()
)

func ColourJSON(l string, p colour.Palette) (string, error) {
	fm := featuremap.NewWithPalette(p)
	err := fm.UnmarshalJSON([]byte(l))
	if err != nil {
		return "", microerror.Mask(err)
	}

	b, err := fm.MarshalJSON()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(b), nil
}

func ColourText(l string, p colour.Palette) (string, error) {
	var values []string
	for _, v := range strings.Split(l, " ") {
		values = append(values, p.Value(v))
	}

	return strings.Join(values, " "), nil
}

func Fields(l string, fields []string) (string, error) {
	var expressions []*regexp.Regexp
	for _, f := range fields {
		expressions = append(expressions, regexp.MustCompile(f))
	}

	var m map[string]string
	err := json.Unmarshal([]byte(l), &m)
	if err != nil {
		return "", microerror.Mask(err)
	}

	fm := featuremap.NewWithPalette(colour.NewNoColourPalette())
	for _, e := range expressions {
		for k, v := range m {
			if e.MatchString(k) {
				if k == "time" {
					t, err := time.Parse(timeFormatFrom, v)
					if err != nil {
						return "", microerror.Mask(err)
					}

					v = t.Format(timeFormatTo)
				}

				fm.Set(k, v)
			}
		}
	}

	newFormat, err := json.Marshal(fm)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(newFormat), nil
}

func IsErr(l string) (bool, error) {
	fm := featuremap.NewWithPalette(colour.NewNoColourPalette())
	err := json.Unmarshal([]byte(l), &fm)
	if err != nil {
		return false, microerror.Mask(err)
	}

	isErr := fm.Has("level") && fm.Get("level") == "error"
	isWar := fm.Has("level") && fm.Get("level") == "warning"

	return isErr || isWar, nil
}

func OutputText(l string) (string, error) {
	fm := featuremap.NewWithPalette(colour.NewNoColourPalette())
	err := json.Unmarshal([]byte(l), &fm)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return strings.Join(fm.Values(), "    "), nil
}
