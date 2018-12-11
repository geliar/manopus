package matcher

import "regexp"

type RegExpMatcher struct {
	*regexp.Regexp
}

func (m *RegExpMatcher) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var exp string
	err = unmarshal(&exp)
	if err != nil {
		return err
	}
	if exp == "" {
		return ErrRegExpEmpty
	}
	var re *regexp.Regexp
	if re, err = regexp.Compile(exp); err == nil {
		m.Regexp = re
	}
	return err
}

func (m *RegExpMatcher) Match(str string) (matches map[string]interface{}, matched bool) {
	if !m.Regexp.MatchString(str) {
		return nil, false
	}
	matched = true
	matches = make(map[string]interface{})
	results := m.Regexp.FindStringSubmatch(str)
	names := m.Regexp.SubexpNames()
	for i, match := range results {
		if i != 0 {
			matches[names[i]] = match
		}
	}
	return
}
