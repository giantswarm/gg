package formatter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"
)

const (
	timeFormatFrom = "2006-01-02T15:04:05.999999-07:00"
	timeFormatTo   = "15:04:05"
)

const (
	indent     = 0
	valueSep   = ","
	null       = "null"
	startMap   = "{"
	endMap     = "}"
	startArray = "["
	endArray   = "]"
)

var (
	colorKey    = color.New(color.FgHiBlue).SprintFunc()
	colorString = color.New(color.FgGreen).SprintFunc()
	colorError  = color.New(color.FgRed).SprintFunc()
)

func Colour(l string, output string) (string, error) {
	s, err := colour(l, output, colorString)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}

func Error(l string, output string) (string, error) {
	s, err := colour(l, output, colorError)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
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

	om := NewOrderedMap()
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

				om.Set(k, v)
			}
		}
	}

	newFormat, err := json.Marshal(om)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return fmt.Sprintf("%s\n", newFormat), nil
}

func IsErr(l string) (bool, error) {
	om := NewOrderedMap()
	err := json.Unmarshal([]byte(l), &om)
	if err != nil {
		return false, microerror.Mask(err)
	}

	isErr := om.Has("level") && om.Get("level") == "error"
	isWar := om.Has("level") && om.Get("level") == "warning"

	return isErr || isWar, nil
}

func Output(l string, output string) (string, error) {
	if output == "json" {
		return l, nil
	}

	om := NewOrderedMap()
	err := json.Unmarshal([]byte(l), &om)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return strings.Join(om.Values(), "    ") + "\n", nil
}

func colour(l string, output string, colour func(v ...interface{}) string) (string, error) {
	if output == "json" {
		om := NewOrderedMap()
		err := json.Unmarshal([]byte(l), &om)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if om.Len() == 0 {
			return startMap + endMap, nil
		}

		keys := om.Keys()

		var s string

		s += startMap
		s += " "

		for _, key := range keys {
			s += colorKey("\"", key, "\"") + ": "

			b, err := json.Marshal(om.Get(key))
			if err != nil {
				return "", microerror.Mask(err)
			}

			s += colour(string(b))
			s += " "
		}

		s += endMap
		s += "\n"

		return s, nil
	}

	var values []string
	for _, v := range strings.Split(l, " ") {
		values = append(values, colour(v))
	}

	return strings.Join(values, " "), nil
}
