// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"github.com/ksoichiro/fint"
	"github.com/ksoichiro/fint/common"
	"github.com/ksoichiro/fint/modules"
	"os"
	"testing"
)

const (
	EnvTerm                   = "TERM"
	SrcRootObjcNormal         = "testdata/objc/FintExample"
	SrcRootObjcEmpty          = "testdata/objc/FintExample_Empty"
	SrcRootObjcSingleError    = "testdata/objc/FintExample_SingleError"
	SrcRootObjcSymlink        = "testdata/objc/link"
	SrcSingleFile             = "testdata/objc/FintExample/FintExample/FEAppDelegate.m"
	SrcNonExistent            = "testdata/non_existent_file"
	SrcMatchingButNonExistent = "testdata/non_existent_file.m"
	TestReportDir             = "report_test_normal"
	TestReportDirWithSubdir   = "report_test_normal/subdir"
	ConfigDefault             = ".fint"
	ConfigNonExistent         = "non_existent_dir"
	ConfigNoModules           = "testdata/config/no_module"
	ConfigNoModuleConfig      = "testdata/config/no_module_config"
	ConfigNoTarget            = "testdata/config/no_target"
	LintIdObjc                = "objc"
	LocaleDefault             = "en"
	LocaleJa                  = "ja"
	TemplateDefault           = "default"
	ErrorsObjcNormal          = 64
)

func TestExecuteAsCommand(t *testing.T) {
	var err error

	// When using dumb TERM, messages should not be colorized.
	os.Setenv(EnvTerm, "dumb")
	err = fint.ExecuteAsCommand(&common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When using other than dumb TERM, messages should be colorized.
	os.Setenv(EnvTerm, "xterm-256color")
	err = fint.ExecuteAsCommand(&common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When there is only one violation, result message form should be singular.
	err = fint.ExecuteAsCommand(&common.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectError(t, err)

	// When SrcRoot is empty, lint should not be executed.
	err = fint.ExecuteAsCommand(&common.Opt{SrcRoot: "", ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc})
	testExpectErrorWithMessage(t, err, "fint: source directory is required.")
}

func TestExecute(t *testing.T) {
	testExecuteNormalWithReport(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Template: TemplateDefault, Force: true}, ErrorsObjcNormal, true, false)
	testExecuteNormalWithReport(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Template: TemplateDefault, Force: true}, ErrorsObjcNormal, false, true)
	testExecuteNormalWithReport(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDirWithSubdir, Template: TemplateDefault, Force: true}, ErrorsObjcNormal, true, true)
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, ErrorsObjcNormal)
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcEmpty, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, 0)
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, 1)
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcSingleError, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc, Quiet: true}, 1)
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcSymlink, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, ErrorsObjcNormal)
}

func TestExecuteError(t *testing.T) {
	testExecuteError(t, &common.Opt{SrcRoot: "", ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc}, "fint: source directory is required.")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: ""}, "fint: ID of the rule set is required.")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: "", Locale: LocaleDefault, Id: LintIdObjc}, "fint: config directory is required.")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: "foo"}, "fint: no matching target to [foo]")
	testExecuteNormalWithReport(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Template: TemplateDefault, Force: true}, ErrorsObjcNormal, true, false)
	testExecuteErrorWithReport(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleDefault, Id: LintIdObjc, Html: TestReportDir, Template: TemplateDefault},
		"fint: report directory already exists. use `-f` option to force reporting.",
		false, true)
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigNonExistent, Locale: LocaleDefault, Id: LintIdObjc}, "stat non_existent_dir: no such file or directory")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigNoModules, Locale: LocaleDefault, Id: LintIdObjc}, "fint: modules directory not found in [testdata/config/no_module/builtin/modules]")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigNoModuleConfig, Locale: LocaleDefault, Id: LintIdObjc}, "open testdata/config/no_module_config/builtin/modules/pattern_match/config.json: no such file or directory")
	testExecuteError(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigNoTarget, Locale: LocaleDefault, Id: LintIdObjc}, "fint: no matching target to ["+LintIdObjc+"]")
}

func TestSetbufsizeAndLint(t *testing.T) {
	// normal
	var m common.Module
	m.Pattern = ".*\\.(m|mm|h)$"
	m.Rules = []common.Rule{
		common.Rule{Id: "ExceedMaxLength", Args: []interface{}{".*", 80.0}, Message: map[string]string{"en": "Line length exceeds %d characters"}}}

	// When the bufSize is set to 0, default size will be set.
	fint.Setbufsize(0)
	_, err := modules.LintWalk(SrcRootObjcNormal, m, LocaleDefault, false, modules.LintMaxLengthFunc)
	if err != nil {
		t.Errorf("Unexpected error occurred: %v", err)
	}

	fint.Setbufsize(1)
	_, err = modules.LintWalk(SrcRootObjcNormal, m, LocaleDefault, false, modules.LintMaxLengthFunc)

	// Do normal test to initialize opt
	testExecuteNormal(t, &common.Opt{SrcRoot: SrcRootObjcNormal, ConfigPath: ConfigDefault, Locale: LocaleJa, Id: LintIdObjc}, ErrorsObjcNormal)
}

func TestCopyDir(t *testing.T) {
	fint.CopyDir("testdata", "testdata_copy")
	os.RemoveAll("testdata_copy")
}

func TestCopyFile(t *testing.T) {
	err := fint.CopyFile(".fint/templates/non_existent_file", "")
	testExpectErrorWithMessage(t, err, "open .fint/templates/non_existent_file: no such file or directory")

	err = fint.CopyFile(".fint/builtin/modules/max_length/config.json", "non_existent_dir/config.json")
	testExpectErrorWithMessage(t, err, "open non_existent_dir/config.json: no such file or directory")
}

func TestClean(t *testing.T) {
	// Remove directories for test
	os.RemoveAll(TestReportDir)
}

func testExecuteNormalWithReport(t *testing.T,
	opt *common.Opt,
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

func testExecuteNormal(t *testing.T, opt *common.Opt, expectedViolations int) {
	v, _ := fint.Execute(opt)
	if len(v) != expectedViolations {
		t.Errorf("Expected violations are [%d] but [%d] found", expectedViolations, len(v))
	}
}

func testExecuteErrorWithReport(
	t *testing.T,
	opt *common.Opt,
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

func testExecuteError(t *testing.T, opt *common.Opt, msg string) {
	_, err := fint.Execute(opt)
	testExpectErrorWithMessage(t, err, msg)
}

func testExpectSuccess(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expected success but an error occurred")
	}
}

func testExpectError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error but not occurred")
	}
}

func testExpectErrorWithMessage(t *testing.T, err error, msg string) {
	testExpectError(t, err)
	if err == nil {
		t.Errorf("Expected error message [%s] but there was no error", msg)
	} else if err.Error() != msg {
		t.Errorf("Expected error message [%s] but was [%s]", msg, err.Error())
	}
}
