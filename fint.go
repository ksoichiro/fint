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
	DirTemplates                = "templates"
	DirSrc                      = "src"
	DirJs                       = "js"
	DirCss                      = "css"
	FileConfig                  = "config.json"
	HtmlIndex                   = "index.html"
	HtmlTmplIndex               = "_index.html"
	HtmlTmplIndexSrclist        = "_index_srclist.html"
	HtmlTmplSrc                 = "_src.html"
	HtmlTmplSrcSrcline          = "_src_srcline.html"
	HtmlTmplSrcViolationMsg     = "_src_violation_msg.html"
	HtmlTmplSrcViolationMsglist = "_src_violation_msglist.html"
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

type Rule struct {
	Pattern string
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

type Config struct {
	RuleSets []RuleSet
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
	pathTmpl := filepath.Join(opt.ConfigPath, DirTemplates, opt.Template)
	CopyFile(filepath.Join(pathTmpl, HtmlTmplIndex), filepath.Join(opt.Html, HtmlIndex))
	CopyFile(filepath.Join(pathTmpl, DirJs, "src.js"), filepath.Join(opt.Html, DirJs, "src.js"))
	CopyFile(filepath.Join(pathTmpl, DirCss, "main.css"), filepath.Join(opt.Html, DirCss, "main.css"))
	CopyFile(filepath.Join(pathTmpl, DirCss, "index.css"), filepath.Join(opt.Html, DirCss, "index.css"))
	CopyFile(filepath.Join(pathTmpl, DirCss, "src.css"), filepath.Join(opt.Html, DirCss, "src.css"))
}

func printReportBody(filename string, vs []Violation, vmap map[int][]Violation) {
	if opt.Html == "" {
		return
	}
	MkReportDir(true)

	// Add source file entry to index
	f, _ := os.OpenFile(filepath.Join(opt.Html, HtmlTmplIndexSrclist), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	srclistTempate, _ := ioutil.ReadFile(filepath.Join(opt.ConfigPath, DirTemplates, opt.Template, HtmlTmplIndexSrclist))
	srclist := string(srclistTempate)
	srclist = replaceTag(srclist, "@SRCPATH@", filename)
	srclist = replaceTag(srclist, "@VIOLATIONS@", fmt.Sprintf("%d", len(vs)))

	f.WriteString(srclist + newlineDefault)
	f.Close()

	var rootPath = ".." + string(filepath.Separator)
	for c := 0; c < strings.Count(filename, string(filepath.Separator)); c++ {
		rootPath = rootPath + ".." + string(filepath.Separator)
	}
	rootPath = strings.TrimSuffix(rootPath, string(filepath.Separator))

	os.MkdirAll(filepath.Join(opt.Html, DirSrc, filepath.Dir(filename)), 0777)

	pathDetail := filepath.Join(opt.Html, DirSrc, filename+".html")
	CopyFile(filepath.Join(opt.ConfigPath, DirTemplates, opt.Template, HtmlTmplSrc), pathDetail)
	replaceTagInFile(pathDetail, "@ROOTPATH@", rootPath)
	replaceTagInFile(pathDetail, "@SRCFILE@", filename)

	pathDetailSrcline := pathDetail + ".srcline.tmp"
	fsrcline, _ := os.OpenFile(pathDetailSrcline, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fsrcline.Close()

	fsrc, _ := os.Open(filename)
	defer fsrc.Close()
	r := bufio.NewReaderSize(fsrc, bufSize)
	for n := 1; true; n++ {
		lineBytes, _, err := r.ReadLine()
		line := string(lineBytes)
		var vsclass string
		var vs []Violation
		var ok bool
		if vs, ok = vmap[n]; ok && 0 < len(vs) {
			vsclass = "violation"
		} else {
			vsclass = "ok"
		}
		srclineBase, _ := ioutil.ReadFile(filepath.Join(opt.ConfigPath, DirTemplates, opt.Template, HtmlTmplSrcSrcline))

		srclineBaseReplaced := replaceTag(string(srclineBase), "@MARKER_CLASS@", vsclass)
		var hasViolations string
		if 0 < len(vs) {
			hasViolations = "true"
		} else {
			hasViolations = "false"
		}
		srclineBaseReplaced = replaceTag(srclineBaseReplaced, "@HAS_VIOLATIONS@", hasViolations)
		srclineBaseReplaced = replaceTag(srclineBaseReplaced, "@LINE@", fmt.Sprintf("%d", n))
		srclineBaseReplaced = replaceTag(srclineBaseReplaced, "@CODE@", line)
		fsrcline.WriteString(srclineBaseReplaced + newlineDefault)

		if 0 < len(vs) {
			msglistBase, _ := ioutil.ReadFile(filepath.Join(opt.ConfigPath, DirTemplates, opt.Template, HtmlTmplSrcViolationMsglist))
			msglistBaseReplaced := replaceTag(string(msglistBase), "@LINE@", fmt.Sprintf("%d", n))

			pathDetailMsg := pathDetail + ".msg.tmp"
			fmsg, _ := os.OpenFile(pathDetailMsg, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			defer fmsg.Close()
			msgBase, _ := ioutil.ReadFile(filepath.Join(opt.ConfigPath, DirTemplates, opt.Template, HtmlTmplSrcViolationMsg))
			for i := range vs {
				msg := replaceTag(string(msgBase), "@VIOLATION_MSG@", vs[i].Message)
				fmsg.WriteString(msg + newlineDefault)
			}
			fmsg.Close()

			msgTemp, _ := ioutil.ReadFile(pathDetailMsg)
			msglistBaseReplaced = replaceTag(msglistBaseReplaced, "@VIOLATION_MSGLIST@", string(msgTemp))
			os.Remove(pathDetailMsg)

			fsrcline.WriteString(msglistBaseReplaced + newlineDefault)
		}
		if err == io.EOF {
			break
		}
	}
	fsrc.Close()
	fsrcline.Close()

	srclineTemp, _ := ioutil.ReadFile(pathDetailSrcline)
	replaceTagInFile(pathDetail, "@SRCLINES@", string(srclineTemp))
	os.Remove(pathDetailSrcline)
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
		lineBytes, _, err := r.ReadLine()
		line := string(lineBytes)

		tagRegexp, _ := regexp.Compile(tag)
		lineReplaced := tagRegexp.ReplaceAllString(line, repl)
		ftmp.WriteString(lineReplaced + newlineDefault)

		if err == io.EOF {
			break
		}
	}
	fin.Close()
	ftmp.Close()
	CopyFile(filename+".tmp", filename)
	os.Remove(filename + ".tmp")
}

func finishReportFiles() {
	if opt.Html == "" {
		return
	}
	srclistTemp, _ := ioutil.ReadFile(filepath.Join(opt.Html, HtmlTmplIndexSrclist))
	replaceTagInFile(filepath.Join(opt.Html, HtmlIndex), "@SRCLIST@", string(srclistTemp))
	os.Remove(filepath.Join(opt.Html, HtmlTmplIndexSrclist))
}

func LoadConfig(file []byte) *Config {
	var c Config
	json.Unmarshal(file, &c)
	return &c
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
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Pattern, line); matched {
						v = Violation{Filename: filename, Line: n, Message: rs.Modules[i].Rules[j].Message[opt.Locale]}
						lvs = append(lvs, v)
						vs = append(vs, v)
					}
				}
			case "max_length":
				for j := range rs.Modules[i].Rules {
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Pattern, line); matched {
						max_len := int(rs.Modules[i].Rules[j].Args[0].(float64))
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

func findRuleSet() (rs RuleSet, err error) {
	for i := range config.RuleSets {
		r := config.RuleSets[i]
		if r.Id == opt.Id {
			rs = r
		}
	}
	if rs.Id == "" {
		err = newError("no matching ruleset to [" + opt.Id + "]")
	}
	return
}

func CheckFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	rs, errRs := findRuleSet()
	if errRs != nil {
		return errRs
	}

	if matched, _ := regexp.MatchString(rs.Pattern, path); matched {
		v, e := CheckSourceFile(path, rs)
		if e != nil {
			return e
		}
		violations = append(violations, v...)
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

	var conf []byte
	if opt.ConfigPath == "" {
		err = newError("config directory is required.")
		return
	}
	conf, err = ioutil.ReadFile(filepath.Join(opt.ConfigPath, FileConfig))
	if err != nil {
		return
	}
	config = LoadConfig(conf)

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
