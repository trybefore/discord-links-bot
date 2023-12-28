package replacer

import "regexp"

var _ Replacer = (*Simple)(nil)

// Simple replaces all matches in *regexp.Regexp regex with replacement
type Simple struct {
	regex       *regexp.Regexp
	replacement string
	name        string
}

func (r *Simple) Name() string {
	return r.name
}

func NewSimple(name string, regex *regexp.Regexp, replacement string) *Simple {
	return &Simple{regex: regex, replacement: replacement, name: name}
}

func (r *Simple) Replace(msg string) (string, error) {
	if !r.Matches(msg) {
		return msg, ErrNoMatch
	}

	return replaceMatches(r.regex, msg, r.replacement), nil
}

func (r *Simple) Matches(msg string) bool {
	return r.regex.MatchString(msg)
}
