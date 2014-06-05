// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"fmt"
	"github.com/ksoichiro/fint/common"
	"regexp"
)

func LintMaxLengthFunc(m common.Module, n int, filename, line, locale string, shouldFix bool) (vs []common.Violation, fixedAny bool, fixedLine string) {
	for i := range m.Rules {
		if matched, _ := regexp.MatchString(m.Rules[i].Args[0].(string), line); matched {
			max_len := int(m.Rules[i].Args[1].(float64))
			if too_long := max_len < len(line); too_long {
				v := common.Violation{Filename: filename, Line: n, Message: fmt.Sprintf(m.Rules[i].Message[locale], max_len)}
				vs = append(vs, v)
			}
		}
	}
	return
}
