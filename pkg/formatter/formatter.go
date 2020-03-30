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

func Colour(l string, output string) (string, error) {
	s, err := colour(l, output, colorString, indentNone, nil)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}

func Error(l string, output string, fields []string) (string, error) {
	s, err := colour(l, output, colorError, indentFour, fields)
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

func colour(l string, output string, colour func(v ...interface{}) string, indent string, fields []string) (string, error) {
	if output == "json" {
		om := NewOrderedMap()
		err := json.Unmarshal([]byte(l), &om)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if om.Len() == 0 {
			return mapStart + mapEnd, nil
		}

		keys := om.Keys()

		var s string

		s += mapStart
		if indent != indentNone {
			s += "\n"
		} else {
			s += " "
		}

		for i, key := range keys {
			l := colorKey("\"", key, "\"") + ": "

			if key == "stack" {
				l += arrayStart + "\n"

				var list []map[string]interface{}
				err := json.Unmarshal([]byte(om.Get(key)), &list)
				if err != nil {
					// TODO all this here is legacy error handling. With microerror
					// changes the stack structure changed. Here we still have to deal
					// with error stacks which are actually not representated as valid
					// JSON. Once all operators make use of the new microerror structures
					// the code below can be removed and replaced with the usual error
					// handling.
					//
					//     if err != nil {
					//         return "", microerror.Mask(err)
					//     }
					//
					// TODO we magically capture error message information from the legacy
					// structures and add an annotation to the printed line. In order to
					// maintain field selection via -f/--field the interface of colour was
					// extended with the fields parameter. Once this code here is removed
					// the fields parameter can be dropped since this is just an ugly hack
					// for now.
					{
						stack := om.Get(key)
						stack = stack[2 : len(stack)-2]
						stack = "[ { \"file\": \"" + stack
						stack = strings.Replace(stack, ".go:", ".go\", \"line\": ", -1)
						stack = strings.Replace(stack, ": } {", "}, {\"file\": \"", -1)
						stack = strings.Replace(stack, "}, {", " }, {", -1)
						stack = strings.Replace(stack, "} {", ", ", -1)

						annotation := regexp.MustCompile(`: [^"][^0-9].*}?`).FindString(stack)
						stack = strings.Replace(stack, annotation, "", -1)
						annotation = annotation[2:len(annotation)]

						var expressions []*regexp.Regexp
						for _, f := range fields {
							expressions = append(expressions, regexp.MustCompile(f))
						}

						for _, e := range expressions {
							if e.MatchString("annotation") {
								s += indent + colorKey("\"annotation\"") + ": " + colour("\""+annotation+"\"") + ",\n"
							}
						}

						stack = stack + " } ]"

						err := json.Unmarshal([]byte(stack), &list)
						if err != nil {
							return "", microerror.Mask(err)
						}
					}
				}

				for j, v := range list {
					l += indent + indent + "{ " + colorKey("\"file\"") + ": " + colour("\""+v["file"].(string)+"\"") + ", " + colorKey("\"line\"") + ": " + colour(v["line"].(float64)) + " }"

					if j+1 < len(list) {
						l += ","
					}

					l += "\n"
				}

				l += indent + arrayEnd
			} else {
				b, err := json.Marshal(om.Get(key))
				if err != nil {
					return "", microerror.Mask(err)
				}

				l += colour(string(b))
			}

			if indent != indentNone {
				s += indent + l
				if i+1 < len(keys) {
					s += ","
				}
				s += "\n"
			} else {
				s += l + " "
			}
		}

		s += mapEnd
		s += "\n"

		return s, nil
	}

	var values []string
	for _, v := range strings.Split(l, " ") {
		values = append(values, colour(v))
	}

	return strings.Join(values, " "), nil
}
