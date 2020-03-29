package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/giantswarm/microerror"
)

type Format string

const (
	JSON Format = "json"
	None Format = "none"
	Text Format = "text"
)

const (
	timeFormatFrom = "2006-01-02T15:04:05.999999-07:00"
	timeFormatTo   = "15:04:05"
)

const (
	indent       = 0
	initialDepth = 0
	valueSep     = ","
	null         = "null"
	startMap     = "{"
	endMap       = "}"
	startArray   = "["
	endArray     = "]"
)

var (
	colorKey    = color.New(color.FgHiBlue)
	colorString = color.New(color.FgGreen)
	colorBool   = color.New(color.FgYellow)
	colorNumber = color.New(color.FgCyan)
	colorNull   = color.New(color.FgMagenta)
)

func Colour(l string, output string) (string, error) {
	if output == "json" {
		newMap := NewOrderedMap()
		err := json.Unmarshal([]byte(l), &newMap)
		if err != nil {
			return "", microerror.Mask(err)
		}

		buffer := bytes.Buffer{}
		marshalValue(newMap, &buffer, initialDepth)

		return buffer.String() + "\n", nil
	}

	var values []string
	for _, v := range strings.Split(l, " ") {
		values = append(values, sprintColor(colorString, v))
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

	newMap := NewOrderedMap()
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

				newMap.Set(k, v)
			}
		}
	}

	newFormat, err := json.Marshal(newMap)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return fmt.Sprintf("%s\n", newFormat), nil
}

func Is(l string) Format {
	if len(l) == 0 {
		return None
	}
	if l[0] == '{' {
		return JSON
	}

	return Text
}

func Output(l string, output string) (string, error) {
	if output == "json" {
		return l, nil
	}

	newMap := NewOrderedMap()
	err := json.Unmarshal([]byte(l), &newMap)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return strings.Join(newMap.Values(), "    ") + "\n", nil
}

func marshalArray(a []interface{}, buf *bytes.Buffer, depth int) {
	if len(a) == 0 {
		buf.WriteString(startArray + endArray)
		return
	}

	buf.WriteString(startArray)
	writeObjSep(buf)

	for i, v := range a {
		writeIndent(buf, depth+1)
		marshalValue(v, buf, depth+1)
		if i < len(a)-1 {
			buf.WriteString(valueSep)
		}
		writeObjSep(buf)
	}
	writeIndent(buf, depth)
	buf.WriteString(endArray)
}

func marshalMap(m map[string]interface{}, buf *bytes.Buffer, depth int) {
	remaining := len(m)

	if remaining == 0 {
		buf.WriteString(startMap + endMap)
		return
	}

	keys := make([]string, 0)
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	buf.WriteString(startMap)
	writeObjSep(buf)

	for _, key := range keys {
		writeIndent(buf, depth+1)
		buf.WriteString(colorKey.Sprintf("\"%s\"", key) + ": ")
		marshalValue(m[key], buf, depth+1)
		remaining--
		if remaining != 0 {
			buf.WriteString(valueSep)
		}
		writeObjSep(buf)
	}
	writeIndent(buf, depth)
	buf.WriteString(endMap)
}

func marshalOrderedMap(om *OrderedMap, buf *bytes.Buffer, depth int) {
	remaining := om.Len()

	if remaining == 0 {
		buf.WriteString(startMap + endMap)
		return
	}

	keys := om.Keys()

	buf.WriteString(startMap)
	writeObjSep(buf)

	for _, key := range keys {
		writeIndent(buf, depth+1)
		buf.WriteString(colorKey.Sprintf("\"%s\"", key) + ": ")
		marshalValue(om.Get(key), buf, depth+1)
		remaining--
		if remaining != 0 {
			buf.WriteString(valueSep)
		}
		writeObjSep(buf)
	}
	writeIndent(buf, depth)
	buf.WriteString(endMap)
}

func marshalString(str string, buf *bytes.Buffer) {
	strBytes, _ := json.Marshal(str)
	str = string(strBytes)

	buf.WriteString(sprintColor(colorString, str))
}

func marshalValue(val interface{}, buf *bytes.Buffer, depth int) {
	switch v := val.(type) {
	case *OrderedMap:
		marshalOrderedMap(v, buf, depth)
	case map[string]interface{}:
		marshalMap(v, buf, depth)
	case []interface{}:
		marshalArray(v, buf, depth)
	case string:
		marshalString(v, buf)
	case float64:
		buf.WriteString(sprintColor(colorNumber, strconv.FormatFloat(v, 'f', -1, 64)))
	case bool:
		buf.WriteString(sprintColor(colorBool, (strconv.FormatBool(v))))
	case nil:
		buf.WriteString(sprintColor(colorNull, null))
	}
}

func sprintColor(c *color.Color, s string) string {
	return c.SprintFunc()(s)
}

func writeIndent(buf *bytes.Buffer, depth int) {
	buf.WriteString(strings.Repeat(" ", indent*depth))
}

func writeObjSep(buf *bytes.Buffer) {
	if indent != 0 {
		buf.WriteByte('\n')
	} else {
		buf.WriteByte(' ')
	}
}
