// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"github.com/ksoichiro/fint"
	"testing"
)

func TestExecute(t *testing.T) {
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: "conf/config.json", Locale: "default", Id: "objc"}, 20)
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: "conf/config.json", Locale: "ja", Id: "objc"}, 20)
}

func TestExecuteError(t *testing.T) {
	testExecuteError(t, &fint.Opt{SrcRoot: "", ConfigPath: "conf/config.json", Locale: "default", Id: "objc"},
		"fint: source directory is required.")
}

func testExecuteNormal(t *testing.T, opt *fint.Opt, expectedViolations int) {
	v, _ := fint.Execute(opt)
	if len(v) != expectedViolations {
		t.Errorf("Expected violations are [%d] but [%d] found", expectedViolations, len(v))
	}
}

func testExecuteError(t *testing.T, opt *fint.Opt, msg string) {
	_, err := fint.Execute(opt)
	if err == nil {
		t.Errorf("Expected error but not occurred")
		return
	}
	if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}
