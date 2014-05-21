// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"errors"
	"github.com/ksoichiro/fint"
	"os"
	"testing"
)

const (
	EnvTerm                 = "TERM"
	SrcRootObjcNormal       = "testdata/objc/FintExample"
	SrcRootObjcEmpty        = "testdata/objc/FintExample_Empty"
	SrcRootObjcSingleError  = "testdata/objc/FintExample_SingleError"
	SrcSingleFile           = "testdata/objc/FintExample/FintExample/FEAppDelegate.m"
	SrcNonExistent          = "testdata/non_existent_file"
	TestReportDir           = "report_test_normal"
	TestReportDirWithSubdir = "report_test_normal/subdir"
	ConfigDefault           = ".fint.json"
	LintIdObjc              = "objc"
	LocaleDefault           = "default"
	LocaleJa                = "ja"
	ErrorsObjcNormal        = 22
)

func TestExecuteAsCommand(t *testing.T) {
	var err error

	// When using dumb TERM, messages should not be colorized.
	os.Setenv(EnvTerm, "dumb")
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When using other than dumb TERM, messages should be colorized.
	os.Setenv(EnvTerm, "xterm-256color")
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When there is only one violation, result message form should be singular.
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When SrcRoot is empty, lint should not be executed.
	err = fint.ExecuteAsCommand(&fint.Opt{SrcRoot: "", ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectErrorWithMessage(t, err, "fint: source directory is required.")
}

func TestExecute(t *testing.T) {
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Force: true}, ErrorsObjcNormal, true, false)
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Force: true}, ErrorsObjcNormal, false, true)
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDirWithSubdir, Force: true}, ErrorsObjcNormal, true, true)
	testExecuteNormal(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, ErrorsObjcNormal)
	testExecuteNormal(t, &fint.Opt{SrcRoot: SrcRootObjcEmpty, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, 0)
	testExecuteNormal(t, &fint.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, 1)
	testExecuteNormal(t, &fint.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc, Quiet: true}, 1)
}

func TestExecuteError(t *testing.T) {
	testExecuteError(t, &fint.Opt{SrcRoot: "", ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc}, "fint: source directory is required.")
	testExecuteError(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: ""}, "fint: ID of the rule set is required.")
	testExecuteError(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: "", Locale: LocaleDefault, Id: LintIdObjc}, "open : no such file or directory")
	testExecuteError(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: "foo"}, "fint: no matching ruleset to [foo]")
	testExecuteNormalWithReport(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Force: true}, ErrorsObjcNormal, true, false)
	testExecuteErrorWithReport(t, &fint.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir},
		"fint: report directory already exists. use `-f` option to force reporting.",
		false, true)
}

func TestCheckSourceFile(t *testing.T) {
	_, err := fint.CheckSourceFile(SrcNonExistent, fint.RuleSet{})
	testExpectErrorWithMessage(t, err, "fint: cannot open "+SrcNonExistent)
}

func TestCheckFile(t *testing.T) {
	// Pass an error message as filepath.Walk will do.
	errIn := errors.New("test message")
	f, _ := os.Stat(".")
	err := fint.CheckFile("", f, errIn)
	testExpectErrorWithMessage(t, err, errIn.Error())
}

func TestSetbufsize(t *testing.T) {
	// When the bufSize is set to 0, default size will be set.
	fint.Setbufsize(0)
	_, err := fint.CheckSourceFile(SrcSingleFile, fint.RuleSet{})
	if err != nil {
		t.Errorf("Unexpected error occurred: %v", err)
	}

	fint.Setbufsize(1)
	_, err = fint.CheckSourceFile(SrcSingleFile, fint.RuleSet{})
	testExpectErrorWithMessage(t, err, "fint: too long line: "+SrcSingleFile)
}

func TestClean(t *testing.T) {
	// Remove directories for test
	os.RemoveAll(TestReportDir)
}

func testExecuteNormalWithReport(t *testing.T,
	opt *fint.Opt,
	expectedViolations int,
	removeReportBefore bool,
	removeReportAfter bool) {
	// Ensure that the report directory does not exist
	if removeReportBefore {
		os.RemoveAll(opt.Html)
	}

	testExecuteNormal(t, opt, expectedViolations)

	// Remove report directory
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
	// Ensure that the report directory does not exist
	if removeReportBefore {
		os.RemoveAll(opt.Html)
	}

	testExecuteError(t, opt, msg)

	// Remove report directory
	if removeReportAfter {
		os.RemoveAll(opt.Html)
	}
}

func testExecuteError(t *testing.T, opt *fint.Opt, msg string) {
	_, err := fint.Execute(opt)
	testExpectErrorWithMessage(t, err, msg)
}

func testExpectError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}
}

func testExpectErrorWithMessage(t *testing.T, err error, msg string) {
	testExpectError(t, err)
	if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}
