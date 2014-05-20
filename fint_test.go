// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"errors"
	"github.com/ksoichiro/fint"
	"os"
	"testing"
)

func TestExecuteAsCommand(t *testing.T) {
	var err error
	os.Setenv("TERM", "dumb")
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc"})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}

	os.Setenv("TERM", "xterm-256color")
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc"})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}

	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: "testdata/objc/FintExample_SingleError", ConfigPath: ".fint.json", Locale: "default", Id: "objc"})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}

	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: "", ConfigPath: ".fint.json", Locale: "default", Id: "objc"})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}
	msg := "fint: source directory is required."
	if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}

func TestExecute(t *testing.T) {
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc", Html: "report_test_normal", Force: true}, 20, true, false)
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc", Html: "report_test_normal", Force: true}, 20, false, true)
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc", Html: "report_test_normal/subdir", Force: true}, 20, true, true)
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "ja", Id: "objc"}, 20)
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample_Empty", ConfigPath: ".fint.json", Locale: "ja", Id: "objc"}, 0)
	testExecuteNormal(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample_SingleError", ConfigPath: ".fint.json", Locale: "ja", Id: "objc"}, 1)
}

func TestExecuteError(t *testing.T) {
	testExecuteError(t, &fint.Opt{SrcRoot: "", ConfigPath: ".fint.json", Locale: "default", Id: "objc"},
		"fint: source directory is required.")
	testExecuteError(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: ""},
		"fint: ID of the rule set is required.")
	testExecuteError(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: "", Locale: "default", Id: "objc"},
		"open : no such file or directory")
	testExecuteError(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "foo"},
		"fint: no matching ruleset to [foo]")
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc", Html: "report_test_normal", Force: true}, 20, true, false)
	testExecuteErrorWithReport(t, &fint.Opt{SrcRoot: "testdata/objc/FintExample", ConfigPath: ".fint.json", Locale: "default", Id: "objc", Html: "report_test_normal"},
		"fint: report directory already exists. use `-f` option to force reporting.",
		false, true)
}

func TestCheckSourceFile(t *testing.T) {
	filename := "testdata/non_existent_file"
	_, err := fint.CheckSourceFile(filename, fint.RuleSet{})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}
	msg := "fint: cannot open " + filename
	if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}

func TestCheckFile(t *testing.T) {
	errIn := errors.New("test message")
	f, _ := os.Stat(".")
	err := fint.CheckFile("", f, errIn)
	if err.Error() != errIn.Error() {
		t.Errorf("Expected error message [%s] but was [%s]", errIn.Error(), err.Error())
	}
}

func TestSetbufsize(t *testing.T) {
	var (
		filename = "testdata/objc/FintExample/FintExample/FEAppDelegate.m"
		msg      string
	)
	fint.Setbufsize(0)
	_, err := fint.CheckSourceFile(filename, fint.RuleSet{})
	if err != nil {
		t.Errorf("Unexpected error occurred: %v", err)
	}

	fint.Setbufsize(1)
	_, err = fint.CheckSourceFile(filename, fint.RuleSet{})
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}
	msg = "fint: too long line: " + filename
	if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}

func TestClean(t *testing.T) {
	os.RemoveAll("report_test_normal")
}

func testExecuteNormalWithReport(t *testing.T,
	opt *fint.Opt,
	expectedViolations int,
	removeReportBefore bool,
	removeReportAfter bool) {
	if removeReportBefore {
		os.RemoveAll(opt.Html)
	}
	testExecuteNormal(t, opt, expectedViolations)
	if removeReportAfter {
		os.RemoveAll(opt.Html)
	}
}

func testExecuteNormal(t *testing.T, opt *fint.Opt, expectedViolations int) {
	v, _ := fint.Execute(opt)
	if len(v) != expectedViolations {
		t.Errorf("Expected violations are [%d] but [%d] found", expectedViolations, len(v))
	}
}

func testExecuteErrorWithReport(
	t *testing.T,
	opt *fint.Opt,
	msg string,
	removeReportBefore bool,
	removeReportAfter bool) {
	if removeReportBefore {
		os.RemoveAll(opt.Html)
	}
	testExecuteError(t, opt, msg)
	if removeReportAfter {
		os.RemoveAll(opt.Html)
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
