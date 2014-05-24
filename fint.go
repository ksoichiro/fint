// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	errPrefix                   = "fint: "
	defaultBufSize              = 4096
	newlineDefault              = "\r\n"
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

type Opt struct {
	SrcRoot    string
	ConfigPath string
	Locale     string
	Id         string
	Html       string
	Template   string
	Force      bool
	Quiet      bool
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
	Id    string
	Rules []Rule
}

type RuleSet struct {
	Id          string
	Description string
	Pattern     string
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
}

var (
	opt        *Opt
	config     *Config
	violations []Violation
	term       string
	bufSize    int
)

func CopyFile(src, dst string) (err error) {
	fin, err := os.Open(src)
	if err != nil {
		return
	}
	defer fin.Close()
	fout, err := os.Create(dst)
	if err != nil {
		return
	}
	defer fout.Close()
	if _, err = io.Copy(fout, fin); err != nil {
		return
	}
	err = fout.Sync()
	return
}

func CopyDir(src, dst string) {
	fi, _ := os.Stat(src)
	fis, _ := ioutil.ReadDir(src)
	os.MkdirAll(dst, fi.Mode())
	for i := range fis {
		entry := fis[i]
		entryPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, filepath.Base(entry.Name()))
		if entry.IsDir() {
			CopyDir(entryPath, dstPath)
		} else {
			CopyFile(entryPath, dstPath)
		}
	}
}

func newError(message string) error {
	return errors.New(errPrefix + message)
}

func printViolation(v Violation) {
	var format string
	if term == "dumb" {
		format = "%s:%d:1: warning: %s\n"
	} else {
		format = "[1;37m%s:%d:1: [1;35mwarning:[1;37m %s[m\n"
	}
	fmt.Printf(format, v.Filename, v.Line, v.Message)
}

func printReportHeader() {
	if opt.Html == "" {
		return
	}
	os.MkdirAll(filepath.Join(opt.Html, DirJs), 0777)
	os.MkdirAll(filepath.Join(opt.Html, DirCss), 0777)
	pathTmpl := filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template)
	CopyFile(filepath.Join(pathTmpl, HtmlTmplIndex), filepath.Join(opt.Html, HtmlIndex))
	CopyDir(filepath.Join(pathTmpl, DirJs), filepath.Join(opt.Html, DirJs))
	CopyDir(filepath.Join(pathTmpl, DirCss), filepath.Join(opt.Html, DirCss))
}

func printReportBody(filename string, vs []Violation, vmap map[int][]Violation) {
	if opt.Html == "" {
		return
	}
	MkReportDir(true)

	// Add source file entry to index
	f, _ := os.OpenFile(filepath.Join(opt.Html, HtmlTmplIndexSrclist), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	srclist := readFile(filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template, HtmlTmplIndexSrclist))
	srclist = replaceTag(srclist, TagSrcPath, filename)
	srclist = replaceTag(srclist, TagViolations, fmt.Sprintf("%d", len(vs)))

	f.WriteString(srclist + newlineDefault)
	f.Close()

	var rootPath = ".." + string(filepath.Separator)
	for c := 0; c < strings.Count(filename, string(filepath.Separator)); c++ {
		rootPath = rootPath + ".." + string(filepath.Separator)
	}
	rootPath = strings.TrimSuffix(rootPath, string(filepath.Separator))

	os.MkdirAll(filepath.Join(opt.Html, DirSrc, filepath.Dir(filename)), 0777)

	pathDetail := filepath.Join(opt.Html, DirSrc, filename+".html")
	CopyFile(filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template, HtmlTmplSrc), pathDetail)
	replaceTagInFile(pathDetail, TagRootPath, rootPath)
	replaceTagInFile(pathDetail, TagSrcPath, filename)

	pathDetailSrcline := pathDetail + ".srcline.tmp"
	fsrcline, _ := os.OpenFile(pathDetailSrcline, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fsrcline.Close()

	fsrc, _ := os.Open(filename)
	defer fsrc.Close()
	r := bufio.NewReaderSize(fsrc, bufSize)
	for n := 1; true; n++ {
		line, _, err := readLine(r)
		var markerCls string
		var vs []Violation
		var ok bool
		if vs, ok = vmap[n]; ok && 0 < len(vs) {
			markerCls = CssMarkerClsNg
		} else {
			markerCls = CssMarkerClsOk
		}

		srcline := readFile(filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template, HtmlTmplSrcSrcline))
		srcline = replaceTag(string(srcline), TagMarkerClass, markerCls)
		var hasViolations string
		if 0 < len(vs) {
			hasViolations = "true"
		} else {
			hasViolations = "false"
		}
		srcline = replaceTag(srcline, TagHasViolations, hasViolations)
		srcline = replaceTag(srcline, TagLineNumber, fmt.Sprintf("%d", n))
		srcline = replaceTag(srcline, TagCode, line)
		fsrcline.WriteString(srcline + newlineDefault)

		if 0 < len(vs) {
			msglist := replaceTag(readFile(filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template, HtmlTmplSrcViolationMsglist)), TagLineNumber, fmt.Sprintf("%d", n))

			pathDetailMsg := pathDetail + ".msg.tmp"
			fmsg, _ := os.OpenFile(pathDetailMsg, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer fmsg.Close()
			msgTmpl := readFile(filepath.Join(opt.ConfigPath, DirBuiltin, DirTemplates, opt.Template, HtmlTmplSrcViolationMsg))
			for i := range vs {
				msg := replaceTag(msgTmpl, TagViolationMsg, vs[i].Message)
				fmsg.WriteString(msg + newlineDefault)
			}
			fmsg.Close()

			msglist = replaceTag(msglist, TagViolationMsglist, readFile(pathDetailMsg))
			os.Remove(pathDetailMsg)

			fsrcline.WriteString(msglist + newlineDefault)
		}
		if err == io.EOF {
			break
		}
	}
	fsrc.Close()
	fsrcline.Close()

	replaceTagInFile(pathDetail, TagSrclines, readFile(pathDetailSrcline))
	os.Remove(pathDetailSrcline)
}

func readFile(filename string) string {
	content, _ := ioutil.ReadFile(filename)
	return string(content)
}

func replaceTag(s, tag, repl string) string {
	exp, _ := regexp.Compile(tag)
	return exp.ReplaceAllString(s, repl)
}

func replaceTagInFile(filename, tag, repl string) {
	fin, _ := os.Open(filename)
	defer fin.Close()
	ftmp, _ := os.OpenFile(filename+".tmp", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer ftmp.Close()

	r := bufio.NewReaderSize(fin, bufSize)
	for n := 1; true; n++ {
		line, _, err := readLine(r)

		tagRegexp, _ := regexp.Compile(tag)
		line = tagRegexp.ReplaceAllString(line, repl)
		ftmp.WriteString(line + newlineDefault)

		if err == io.EOF {
			break
		}
	}
	fin.Close()
	ftmp.Close()
	CopyFile(filename+".tmp", filename)
	os.Remove(filename + ".tmp")
}

func readLine(r *bufio.Reader) (line string, isPrefix bool, err error) {
	var lineBytes []byte
	lineBytes, isPrefix, err = r.ReadLine()
	line = string(lineBytes)
	return
}

func finishReportFiles() {
	if opt.Html == "" {
		return
	}
	replaceTagInFile(filepath.Join(opt.Html, HtmlIndex), TagSrclist, readFile(filepath.Join(opt.Html, HtmlTmplIndexSrclist)))
	os.Remove(filepath.Join(opt.Html, HtmlTmplIndexSrclist))
}

func dirExists(path string) error {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}
func LoadConfig() (err error) {
	// Get config directory(.fint)
	pathConfig := opt.ConfigPath
	if err = dirExists(pathConfig); err != nil {
		return
	}

	// Get target ID directory
	pathTarget := filepath.Join(pathConfig, DirBuiltin, DirTargets, opt.Id)
	if err = dirExists(pathTarget); err != nil {
		return newError("no matching target to [" + opt.Id + "]")
	}

	// Get modules directory
	pathModules := filepath.Join(pathConfig, DirBuiltin, DirModules)
	if err = dirExists(pathModules); err != nil {
		return newError("modules directory not found in [" + pathModules + "]")
	}

	// Load .fint/builtin/modules/*/config.json
	config = new(Config)
	filesModules, _ := ioutil.ReadDir(pathModules)
	for i := range filesModules {
		entry := filesModules[i]
		if entry.IsDir() {
			// entry name is the name of module
			entryPath := filepath.Join(pathModules, entry.Name())
			var configBytes []byte
			configBytes, err = ioutil.ReadFile(filepath.Join(entryPath, FileConfig))
			if err != nil {
				return
			}

			var c ModuleConfig
			json.Unmarshal(configBytes, &c)

			// Set name as Id to be searchable
			c.Id = entry.Name()
			config.ModuleConfigs = append(config.ModuleConfigs, c)
		}
	}

	// Load target ruleset
	var configBytes []byte
	configBytes, err = ioutil.ReadFile(filepath.Join(pathTarget, FileRuleSet))
	if err != nil {
		return newError("no matching target to [" + opt.Id + "]")
	}
	var target Target
	json.Unmarshal(configBytes, &target)
	config.Targets = append(config.Targets, target)

	// Load target locales
	filesTargetLocales, _ := ioutil.ReadDir(filepath.Join(pathTarget, DirLocales))
	for i := range filesTargetLocales {
		entry := filesTargetLocales[i]
		locale := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		// Get contents of en.json, ja.json, ...
		var configBytes []byte
		configBytes, _ = ioutil.ReadFile(filepath.Join(pathTarget, DirLocales, entry.Name()))

		var lt LocalizedTarget
		json.Unmarshal(configBytes, &lt)

		for i := range target.RuleSets {
			for j := range target.RuleSets[i].Modules {
				for k := range target.RuleSets[i].Modules[j].Rules {
					// Pass all localized messages for each rules
					if target.RuleSets[i].Modules[j].Rules[k].Message == nil {
						target.RuleSets[i].Modules[j].Rules[k].Message = make(map[string]string)
					}
					target.RuleSets[i].Modules[j].Rules[k].Message[locale] = lt.RuleSets[i].Modules[j].Rules[k].Message
				}
			}
		}
	}
	return
}

func Setbufsize(size int) {
	if 0 < size {
		bufSize = size
	}
}

func CheckSourceFile(filename string, rs RuleSet) (vs []Violation, err error) {
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		err = newError("cannot open " + filename)
		return
	}
	defer f.Close()
	if bufSize == 0 {
		bufSize = defaultBufSize
	}
	r := bufio.NewReaderSize(f, bufSize)
	vmap := make(map[int][]Violation)
	for n := 1; true; n++ {
		var (
			lineBytes []byte
			isPrefix  bool
		)
		lineBytes, isPrefix, err = r.ReadLine()
		if isPrefix {
			err = newError(fmt.Sprintf("too long line: %s", filename))
			return
		}
		line := string(lineBytes)
		if err != io.EOF && err != nil {
			return
		}
		var lvs []Violation
		var v Violation
		for i := range rs.Modules {
			switch rs.Modules[i].Id {
			case "pattern_match":
				for j := range rs.Modules[i].Rules {
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Args[0].(string), line); matched {
						v = Violation{Filename: filename, Line: n, Message: rs.Modules[i].Rules[j].Message[opt.Locale]}
						lvs = append(lvs, v)
						vs = append(vs, v)
					}
				}
			case "max_length":
				for j := range rs.Modules[i].Rules {
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Args[0].(string), line); matched {
						max_len := int(rs.Modules[i].Rules[j].Args[1].(float64))
						if too_long := max_len < len(line); too_long {
							v = Violation{Filename: filename, Line: n, Message: fmt.Sprintf(rs.Modules[i].Rules[j].Message[opt.Locale], max_len)}
							lvs = append(lvs, v)
							vs = append(vs, v)
						}
					}
				}
			}
		}
		vmap[n] = lvs
		if err == io.EOF {
			err = nil
			break
		}
	}

	printReportBody(filename, vs, vmap)

	return
}

func CheckFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	target := config.Targets[0]
	for i := range target.RuleSets {
		rs := target.RuleSets[i]
		if matched, _ := regexp.MatchString(rs.Pattern, path); matched {
			v, e := CheckSourceFile(path, rs)
			if e != nil {
				return e
			}
			violations = append(violations, v...)
		}
	}
	return nil
}

func pluralize(value int, singular, plural string) string {
	if value < 2 {
		return singular
	}
	return plural
}

func MkReportDir(whenNotExist bool) (err error) {
	if opt.Html == "" {
		return
	}
	if !whenNotExist {
		if opt.Force {
			os.RemoveAll(opt.Html)
		} else {
			err = newError("report directory already exists. use `-f` option to force reporting.")
			return
		}
	}
	err = os.MkdirAll(opt.Html, 0777)
	return
}

func Execute(o *Opt) (v []Violation, err error) {
	// Clear global vars
	violations = []Violation{}
	bufSize = 0

	if o.SrcRoot == "" {
		err = newError("source directory is required.")
		return
	}
	if o.Id == "" {
		err = newError("ID of the rule set is required.")
		return
	}
	opt = o
	err = MkReportDir(false)
	if err != nil {
		return
	}

	if opt.ConfigPath == "" {
		err = newError("config directory is required.")
		return
	}

	err = LoadConfig()
	if err != nil {
		return
	}

	printReportHeader()

	err = filepath.Walk(opt.SrcRoot, CheckFile)

	finishReportFiles()

	return violations, err
}

func ExecuteAsCommand(o *Opt) (err error) {
	term = os.Getenv("TERM")
	_, err = Execute(o)
	if err != nil {
		if !o.Quiet {
			fmt.Println(err)
		}
		return
	}
	if !o.Quiet {
		for i := range violations {
			printViolation(violations[i])
		}
	}

	if 0 < len(violations) {
		if !o.Quiet {
			fmt.Printf("\n%d %s generated.\n",
				len(violations), pluralize(len(violations), "warning", "warnings"))
		}
		err = newError("error while executing lint")
	}
	return
}
