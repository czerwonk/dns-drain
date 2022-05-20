package undrain

import "regexp"

type Options struct {
	DryRun     bool
	ZoneFilter *regexp.Regexp
	SkipFilter *regexp.Regexp
	Limit      int64
}
