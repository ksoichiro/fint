// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"github.com/ksoichiro/fint/common"
	"regexp"
)

func LintPatternMatchFunc(m common.Module, n int, filename, line, locale string, shouldFix bool) (vs []common.Violation, fixedAny bool, fixedLine string) {
	in := line
	for i := range m.Rules {
		pattern := m.Rules[i].Args[0].(string)
		if matched, _ := regexp.MatchString(pattern, in); matched {
			// Pattern to be excluded
			excludePattern := m.Rules[i].Args[1].(string)
			if excludePattern != "" {
				exp, _ := regexp.Compile(pattern)
				// Exclude this line if the rest string matches to excludePattern
				if matched, _ := regexp.MatchString(excludePattern, exp.FindString(in)); matched {
					continue
				}
			}
			var fixed bool
			var fix string
			if shouldFix && 3 <= len(m.Rules[i].Args) {
				for true {
					// Fix all violations in this line
					repl := m.Rules[i].Args[2].(string)
					exp, _ := regexp.Compile(pattern)
					fix = exp.ReplaceAllString(in, repl)
					if in == fix {
						break
					} else {
						in = fix
						fixed = true
						fixedAny = true
					}
				}
			}
			v := common.Violation{Filename: filename, Line: n, Message: m.Rules[i].Message[locale],
				Fixed: fixed, Fix: fix}
			vs = append(vs, v)
		}
	}
	if in != line {
		fixedLine = in
	}
	return
}
