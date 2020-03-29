package matcher

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"
)

func Match(l string, selects []string) (bool, error) {
	var expressions [][]*regexp.Regexp
	{
		for _, s := range selects {
			split := strings.Split(s, ":")

			var pair []*regexp.Regexp
			pair = append(pair, regexp.MustCompile(split[0]))
			pair = append(pair, regexp.MustCompile(split[1]))
			expressions = append(expressions, pair)
		}
	}

	var m map[string]string
	err := json.Unmarshal([]byte(l), &m)
	if err != nil {
		return false, microerror.Mask(err)
	}

	var matched int

	for _, pair := range expressions {
		matches := pairMatchesMapping(pair, m)
		if matches {
			matched++
		}
	}

	if matched == len(expressions) {
		return true, nil
	}

	return false, nil
}

func Value(l string, s string) (string, error) {
	var m map[string]string
	err := json.Unmarshal([]byte(l), &m)
	if err != nil {
		return "", microerror.Mask(err)
	}

	expression := regexp.MustCompile(s)

	for k, v := range m {
		if expression.MatchString(k) {
			return v, nil
		}
	}

	return "", nil
}

func pairMatchesMapping(pair []*regexp.Regexp, m map[string]string) bool {
	for k, v := range m {
		matches := pair[0].MatchString(k) && pair[1].MatchString(v)
		if matches {
			return true
		}
	}

	return false
}
