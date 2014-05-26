// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ksoichiro/fint/common"
	"github.com/ksoichiro/fint/modules"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	errPrefix      = "fint: "
	defaultBufSize = 4096
	newlineDefault = "\r\n"
)

var (
	opt        *common.Opt
	config     *common.Config
	violations []common.Violation
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

func printViolation(v common.Violation) {
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
	os.MkdirAll(filepath.Join(opt.Html, common.DirJs), 0777)
	os.MkdirAll(filepath.Join(opt.Html, common.DirCss), 0777)
	pathTmpl := filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template)
	CopyFile(filepath.Join(pathTmpl, common.HtmlTmplIndex), filepath.Join(opt.Html, common.HtmlIndex))
	CopyDir(filepath.Join(pathTmpl, common.DirJs), filepath.Join(opt.Html, common.DirJs))
	CopyDir(filepath.Join(pathTmpl, common.DirCss), filepath.Join(opt.Html, common.DirCss))
}

func printReportBody(filename string, vs []common.Violation, vmap map[int][]common.Violation) {
	if opt.Html == "" {
		return
	}
	MkReportDir(true)

	// Add source file entry to index
	f, _ := os.OpenFile(filepath.Join(opt.Html, common.HtmlTmplIndexSrclist), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	srclist := readFile(filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template, common.HtmlTmplIndexSrclist))
	srclist = replaceTag(srclist, common.TagSrcPath, filename)
	srclist = replaceTag(srclist, common.TagViolations, fmt.Sprintf("%d", len(vs)))

	f.WriteString(srclist + newlineDefault)
	f.Close()

	var rootPath = ".." + string(filepath.Separator)
	for c := 0; c < strings.Count(filename, string(filepath.Separator)); c++ {
		rootPath = rootPath + ".." + string(filepath.Separator)
	}
	rootPath = strings.TrimSuffix(rootPath, string(filepath.Separator))

	os.MkdirAll(filepath.Join(opt.Html, common.DirSrc, filepath.Dir(filename)), 0777)

	pathDetail := filepath.Join(opt.Html, common.DirSrc, filename+".html")
	CopyFile(filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template, common.HtmlTmplSrc), pathDetail)
	replaceTagInFile(pathDetail, common.TagRootPath, rootPath)
	replaceTagInFile(pathDetail, common.TagSrcPath, filename)

	pathDetailSrcline := pathDetail + ".srcline.tmp"
	fsrcline, _ := os.OpenFile(pathDetailSrcline, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fsrcline.Close()

	fsrc, _ := os.Open(filename)
	defer fsrc.Close()
	r := bufio.NewReaderSize(fsrc, bufSize)
	for n := 1; true; n++ {
		line, _, err := readLine(r)
		var markerCls string
		var vs []common.Violation
		var ok bool
		if vs, ok = vmap[n]; ok && 0 < len(vs) {
			markerCls = common.CssMarkerClsNg
		} else {
			markerCls = common.CssMarkerClsOk
		}

		srcline := readFile(filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template, common.HtmlTmplSrcSrcline))
		srcline = replaceTag(string(srcline), common.TagMarkerClass, markerCls)
		var hasViolations string
		if 0 < len(vs) {
			hasViolations = "true"
		} else {
			hasViolations = "false"
		}
		srcline = replaceTag(srcline, common.TagHasViolations, hasViolations)
		srcline = replaceTag(srcline, common.TagLineNumber, fmt.Sprintf("%d", n))
		srcline = replaceTag(srcline, common.TagCode, line)
		fsrcline.WriteString(srcline + newlineDefault)

		if 0 < len(vs) {
			msglist := replaceTag(readFile(filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template, common.HtmlTmplSrcViolationMsglist)), common.TagLineNumber, fmt.Sprintf("%d", n))

			pathDetailMsg := pathDetail + ".msg.tmp"
			fmsg, _ := os.OpenFile(pathDetailMsg, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer fmsg.Close()
			msgTmpl := readFile(filepath.Join(opt.ConfigPath, common.DirBuiltin, common.DirTemplates, opt.Template, common.HtmlTmplSrcViolationMsg))
			for i := range vs {
				msg := replaceTag(msgTmpl, common.TagViolationMsg, vs[i].Message)
				fmsg.WriteString(msg + newlineDefault)
			}
			fmsg.Close()

			msglist = replaceTag(msglist, common.TagViolationMsglist, readFile(pathDetailMsg))
			os.Remove(pathDetailMsg)

			fsrcline.WriteString(msglist + newlineDefault)
		}
		if err == io.EOF {
			break
		}
	}
	fsrc.Close()
	fsrcline.Close()

	replaceTagInFile(pathDetail, common.TagSrclines, readFile(pathDetailSrcline))
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
	replaceTagInFile(filepath.Join(opt.Html, common.HtmlIndex), common.TagSrclist, readFile(filepath.Join(opt.Html, common.HtmlTmplIndexSrclist)))
	os.Remove(filepath.Join(opt.Html, common.HtmlTmplIndexSrclist))
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
	pathTarget := filepath.Join(pathConfig, common.DirBuiltin, common.DirTargets, opt.Id)
	if err = dirExists(pathTarget); err != nil {
		return newError("no matching target to [" + opt.Id + "]")
	}

	// Get modules directory
	pathModules := filepath.Join(pathConfig, common.DirBuiltin, common.DirModules)
	if err = dirExists(pathModules); err != nil {
		return newError("modules directory not found in [" + pathModules + "]")
	}

	// Load .fint/builtin/modules/*/config.json
	config = new(common.Config)
	filesModules, _ := ioutil.ReadDir(pathModules)
	for i := range filesModules {
		entry := filesModules[i]
		if entry.IsDir() {
			// entry name is the name of module
			entryPath := filepath.Join(pathModules, entry.Name())
			var configBytes []byte
			configBytes, err = ioutil.ReadFile(filepath.Join(entryPath, common.FileConfig))
			if err != nil {
				return
			}

			var c common.ModuleConfig
			json.Unmarshal(configBytes, &c)

			// Set name as Id to be searchable
			c.Id = entry.Name()
			config.ModuleConfigs = append(config.ModuleConfigs, c)
		}
	}

	// Load target ruleset
	var configBytes []byte
	configBytes, err = ioutil.ReadFile(filepath.Join(pathTarget, common.FileRuleSet))
	if err != nil {
		return newError("no matching target to [" + opt.Id + "]")
	}
	var target common.Target
	json.Unmarshal(configBytes, &target)
	config.Targets = append(config.Targets, target)

	// Load target locales
	filesTargetLocales, _ := ioutil.ReadDir(filepath.Join(pathTarget, common.DirLocales))
	for i := range filesTargetLocales {
		entry := filesTargetLocales[i]
		locale := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		// Get contents of en.json, ja.json, ...
		var configBytes []byte
		configBytes, _ = ioutil.ReadFile(filepath.Join(pathTarget, common.DirLocales, entry.Name()))

		var lt common.LocalizedTarget
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

func CheckSourceFile(filename string, rs common.RuleSet) (vs []common.Violation, err error) {
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
	vmap := make(map[int][]common.Violation)
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
		var lvs []common.Violation
		for i := range rs.Modules {
			var vsr []common.Violation
			switch rs.Modules[i].Id {
			case "pattern_match":
				vsr = modules.LintPatternMatch(rs.Modules[i], n, filename, line, opt.Locale)
			case "max_length":
				vsr = modules.LintMaxLength(rs.Modules[i], n, filename, line, opt.Locale)
			}
			if vsr != nil {
				lvs = append(lvs, vsr...)
				vs = append(vs, vsr...)
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

func Execute(o *common.Opt) (v []common.Violation, err error) {
	// Clear global vars
	violations = []common.Violation{}
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

func ExecuteAsCommand(o *common.Opt) (err error) {
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
