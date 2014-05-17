// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"github.com/ksoichiro/fint"
	"testing"
)

func TestExecute(t *testing.T) {
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: "conf/config.json", Locale: "default", Id: "objc"}, 12)
}

func testExecuteNormal(t *testing.T, opt *fint.Opt, expectedViolations int) {
	v, _ := fint.Execute(opt)
	if len(v) != expectedViolations {
		t.Errorf("Expected violations are %d but %d found", expectedViolations, len(v))
	}
}
