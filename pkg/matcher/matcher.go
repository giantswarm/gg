package matcher

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gg/pkg/featuremap"
)

// Exp is used to transform single key expressions into key-value expressions as
// expected by Match. Single key expressions are given e.g. via -f/--field and
// -g/--group which we want to match for implicitly.
func Exp(list []string) []string {
	var pairs []string

	for _, s := range list {
		split := strings.Split(s, ":")

		if len(split) == 1 {
			pairs = append(pairs, s+":.*")
		}
	}

	return pairs
}

func ExpWithout(list []string, without ...string) []string {
	var filtered []string
	{
		for _, s := range list {
			split := strings.Split(s, ":")

			if ContainsExp(without, s) {
				continue
			}

			if len(split) == 1 {
				filtered = append(filtered, s)
			}
		}
	}

	return Exp(filtered)
}

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

	var matched int
	{
		fm := featuremap.New()
		err := fm.UnmarshalJSON([]byte(l))
		if err != nil {
			return false, microerror.Mask(err)
		}

		for _, pair := range expressions {
			matches := pairMatchesMapping(pair, fm)
			if matches {
				matched++
			}
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

func pairMatchesMapping(pair []*regexp.Regexp, fm *featuremap.FeatureMap) bool {
	f := fm.EntriesIter()
	for {
		kv, ok := f()
		if !ok {
			break
		}

		matchKey := pair[0].MatchString(kv.Key)

		// In case the value is not a string we cannot match it against the given
		// expression. Then we only compare the key and if it matches we select the
		// line.
		s, ok := kv.Value.(string)
		if matchKey && !ok {
			return true
		}

		matchVal := pair[1].MatchString(s)

		if matchKey && matchVal {
			return true
		}
	}

	return false
}

func ContainsExp(fields []string, field string) bool {
	for _, f := range fields {
		if regexp.MustCompile(field).MatchString(f) {
			return true
		}
	}

	return false
}
