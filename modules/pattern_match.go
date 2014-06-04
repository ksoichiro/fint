// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"github.com/ksoichiro/fint/common"
	"regexp"
)

func LintPatternMatchFunc(m common.Module, n int, filename, line, locale string) (vs []common.Violation, fixedAny bool) {
	in := line
	for i := range m.Rules {
		if matched, _ := regexp.MatchString(m.Rules[i].Args[0].(string), in); matched {
			var fixed bool
			var fix string
			if 2 <= len(m.Rules[i].Args) {
				exp, _ := regexp.Compile(m.Rules[i].Args[0].(string))
				fix = exp.ReplaceAllString(in, m.Rules[i].Args[1].(string))
				in = fix
				fixed = true
				fixedAny = true
			}
			v := common.Violation{Filename: filename, Line: n, Message: m.Rules[i].Message[locale],
				Fixed: fixed, Fix: fix}
			vs = append(vs, v)
		}
	}
	return
}
