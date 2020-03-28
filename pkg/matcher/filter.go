package matcher

import (
	"encoding/json"
	"regexp"

	"github.com/giantswarm/microerror"
)

func Match(l string, expressions [][]*regexp.Regexp) (bool, error) {
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

func pairMatchesMapping(pair []*regexp.Regexp, m map[string]string) bool {
	for k, v := range m {
		matches := pair[0].MatchString(k) && pair[1].MatchString(v)
		if matches {
			return true
		}
	}

	return false
}
