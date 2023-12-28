package replacer

import "strings"

var _ Replacer = (*Deferred)(nil)

type Deferred struct {
	replacer Replacer

	deferredReplacer *strings.Replacer
}

func NewDeferred(childReplacer Replacer, stringReplacer *strings.Replacer) *Deferred {
	return &Deferred{replacer: childReplacer, deferredReplacer: stringReplacer}
}

func (d *Deferred) Replace(msg string) (string, error) {
	output, err := d.replacer.Replace(msg)
	if err != nil {
		return output, err
	}

	return d.deferredReplacer.Replace(output), nil
}

func (d *Deferred) Matches(s string) bool {
	return d.replacer.Matches(s)
}

func (d *Deferred) Name() string {
	return d.replacer.Name()
}
