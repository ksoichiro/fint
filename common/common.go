// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package common

import (
	"errors"
)

const (
	errPrefix      = "fint: "
	DefaultBufSize = 4096
	NewlineDefault = "\r\n"

	DirBuiltin                  = "builtin"
	DirModules                  = "modules"
	DirTargets                  = "targets"
	DirLocales                  = "locales"
	DirTemplates                = "templates"
	DirSrc                      = "src"
	DirJs                       = "js"
	DirCss                      = "css"
	FileConfig                  = "config.json"
	FileRuleSet                 = "ruleset.json"
	HtmlIndex                   = "index.html"
	HtmlTmplIndex               = "_index.html"
	HtmlTmplIndexSrclist        = "_index_srclist.html"
	HtmlTmplSrc                 = "_src.html"
	HtmlTmplSrcSrcline          = "_src_srcline.html"
	HtmlTmplSrcViolationMsg     = "_src_violation_msg.html"
	HtmlTmplSrcViolationMsglist = "_src_violation_msglist.html"
	CssMarkerClsNg              = "ng"
	CssMarkerClsOk              = "ok"
	TagSrcPath                  = "@SRCPATH@"
	TagViolations               = "@VIOLATIONS@"
	TagRootPath                 = "@ROOTPATH@"
	TagMarkerClass              = "@MARKER_CLASS@"
	TagHasViolations            = "@HAS_VIOLATIONS@"
	TagLineNumber               = "@LINE@"
	TagCode                     = "@CODE@"
	TagViolationMsg             = "@VIOLATION_MSG@"
	TagViolationMsglist         = "@VIOLATION_MSGLIST@"
	TagSrclines                 = "@SRCLINES@"
	TagSrclist                  = "@SRCLIST@"
)

var (
	BufSize int
)

type Opt struct {
	SrcRoot    string
	ConfigPath string
	Locale     string
	Id         string
	Html       string
	Template   string
	Force      bool
	Quiet      bool
	Fix        bool
}

type LocalizedRule struct {
	Id      string
	Message string
}
type LocalizedModules struct {
	Id    string
	Rules []LocalizedRule
}
type LocalizedRuleSet struct {
	Id      string
	Modules []LocalizedModules
}
type LocalizedTarget struct {
	RuleSets []LocalizedRuleSet
}

type Rule struct {
	Id      string
	Args    []interface{}
	Message map[string]string
}

type Module struct {
	Id      string
	Pattern string
	Rules   []Rule
}

type RuleSet struct {
	Id          string
	Description string
	Modules     []Module
}

type Target struct {
	Id       string
	RuleSets []RuleSet
	Locales  []string
}

type ModuleConfig struct {
	Id          string
	Type        string
	Description string
	Executable  string
	Args        []interface{}
}

type Config struct {
	ModuleConfigs []ModuleConfig
	Targets       []Target
}

type Violation struct {
	Filename string
	Line     int
	Message  string
	Fixed    bool
	Fix      string
}

func NewError(message string) error {
	return errors.New(errPrefix + message)
}
