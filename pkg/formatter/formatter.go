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
	s, err := colour(l, output, colorString, indentNone, nil, false)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return s, nil
}

func Error(l string, output string, fields []string, dropStack bool) (string, error) {
	s, err := colour(l, output, colorError, indentFour, fields, dropStack)
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

func colour(l string, output string, colourFunc func(v ...interface{}) string, indent string, fields []string, dropStack bool) (string, error) {
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

				value := om.Get(key)
				stack, ok := value.(string)
				if ok {
					err := json.Unmarshal([]byte(stack), &list)
					if err != nil {
						// TODO all this here is legacy error handling. With microerror
						// changes the stack structure changed. Here we still have to deal
						// with error stacks which are actually not representated as valid
						// JSON. Once all operators make use of the new microerror structures
						// the stack value should never be a string anymore but always a real
						// JSON object/array.
						//
						// TODO we magically capture error message information from the legacy
						// structures and add an annotation to the printed line. In order to
						// maintain field selection via -f/--field the interface of colour was
						// extended with the fields parameter. Once this code here is removed
						// the fields parameter can be dropped since this is just an ugly hack
						// for now.
						stack = stack[2 : len(stack)-2]
						stack = strings.Replace(stack, `"`, `\"`, -1)
						stack = "[ { \"file\": \"" + stack
						stack = strings.Replace(stack, ".go:", ".go\", \"line\": ", -1)
						stack = strings.Replace(stack, ": } {", "}, {\"file\": \"", -1)
						stack = strings.Replace(stack, "}, {", " }, {", -1)
						stack = strings.Replace(stack, "} {", ", ", -1)

						annotation := regexp.MustCompile(`: [^"][^0-9].*}?`).FindString(stack)
						stack = strings.Replace(stack, annotation, "", -1)

						if annotation != "" {
							annotation = annotation[2:]
						}

						if regexp.MustCompile(`\"line\": [0-9]{1,}$`).MatchString(stack) {
							stack = stack + " } ]"
						} else {
							stack = stack + "\" } ]"
						}

						err := json.Unmarshal([]byte(stack), &list)
						if err != nil {
							return "", microerror.Mask(err)
						}

						for j, v := range list {
							_, ok := v["line"]
							if !ok {
								if annotation == "" {
									annotation = v["file"].(string)
								} else {
									annotation = v["file"].(string) + ", " + annotation
								}
								list = append(list[:j], list[j+1:]...)
								continue
							}
						}

						var withAnn bool
						var expressions []*regexp.Regexp
						for _, f := range fields {
							expressions = append(expressions, regexp.MustCompile(f))
						}
						for _, e := range expressions {
							if e.MatchString("annotation") {
								withAnn = true
							}
						}

						if withAnn || len(fields) == 0 {
							s += indent + colorKey("\"annotation\"") + ": " + colourFunc("\""+annotation+"\"") + ",\n"
						}
					}
				} else {
					om := value.(*OrderedMap)
					b, err := om.MarshalJSON()
					if err != nil {
						return "", microerror.Mask(err)
					}
					err = json.Unmarshal(b, &list)
					if err != nil {
						var jsonErr microerror.JSONError
						err = json.Unmarshal(b, &jsonErr)
						if err != nil {
							return "", microerror.Mask(err)
						}
						b, err := json.Marshal(jsonErr.Stack)
						if err != nil {
							return "", microerror.Mask(err)
						}
						err = json.Unmarshal(b, &list)
						if err != nil {
							return "", microerror.Mask(err)
						}
					}
				}

				// At this point we added the annotation which is all we wanted. If
				// dropStack is true the user wanted the annotation and not the stack.
				// We need the stack to get the annotation so we just found and added
				// the annotation and drop the stack now.
				if dropStack {
					continue
				}

				for j, v := range list {
					l += indent + indent + "{ " + colorKey("\"file\"") + ": " + colourFunc("\""+v["file"].(string)+"\"") + ", " + colorKey("\"line\"") + ": " + colourFunc(v["line"].(float64)) + " }"

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

				l += colourFunc(string(b))
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

		if strings.HasSuffix(s, ",\n") {
			s = s[:len(s)-2] + "\n"
		}

		s += mapEnd
		s += "\n"

		return s, nil
	}

	var values []string
	for _, v := range strings.Split(l, " ") {
		values = append(values, colourFunc(v))
	}

	return strings.Join(values, " "), nil
}
