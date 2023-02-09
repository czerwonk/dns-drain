// SPDX-FileCopyrightText: (c) 2016 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package drain

import "regexp"

type Options struct {
	DryRun     bool
	Force      bool
	ZoneFilter *regexp.Regexp
	SkipFilter *regexp.Regexp
	TypeFilter string
	Limit      int64
}
