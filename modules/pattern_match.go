// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"github.com/ksoichiro/fint/common"
	"regexp"
)

func LintPatternMatchFunc(m common.Module, n int, filename, line, locale string) (vs []common.Violation) {
	for i := range m.Rules {
		if matched, _ := regexp.MatchString(m.Rules[i].Args[0].(string), line); matched {
			v := common.Violation{Filename: filename, Line: n, Message: m.Rules[i].Message[locale]}
			vs = append(vs, v)
		}
	}
	return
}
