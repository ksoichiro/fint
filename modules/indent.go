// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"github.com/ksoichiro/fint/common"
	"regexp"
)

func LintIndentFunc(m common.Module, n int, filename, line, locale string, shouldFix bool) (vs []common.Violation, fixedAny bool, fixedLine string) {
	in := line
	for i := range m.Rules {
		var pattern string
		var replace string
		switch m.Rules[i].Id {
			case "Whitespaces":
				pattern = "\\t"
				replace = ""
				for j := 0; j < int(m.Rules[i].Args[0].(float64)); j++ {
					replace += " "
				}
			default:
				return
		}
		if matched, _ := regexp.MatchString("^" + pattern + "+", in); matched {
			var fixed bool
			var fix string
			if shouldFix {
				patternRepl := "^(" + pattern + "*)(" + pattern + ")([^" + pattern + "]|$)"
				repl := "$1" + replace + "$3"
				var matchedRepl bool
				for true {
					matchedRepl, _ = regexp.MatchString(patternRepl, in)
					if !matchedRepl {
						break
					}
					exp, _ := regexp.Compile(patternRepl)
					fix = exp.ReplaceAllString(in, repl)
					if in != fix {
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
