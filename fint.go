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
	errPrefix      = "fint: "
	defaultBufSize = 4096
	newlineDefault = "\r\n"
)

type Opt struct {
	SrcRoot    string
	ConfigPath string
	Locale     string
	Id         string
	Html       string
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
	CopyFile(opt.ConfigPath+"/templates/default/_index.html", opt.Html+"/index.html")
}

func printReportBody(filename string, vs []Violation, vmap map[int][]Violation) {
	if opt.Html == "" {
		return
	}
	MkReportDir(true)

	// Add source file entry to index
	f, _ := os.OpenFile(opt.Html+"/_index_srclist.html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	srclistTempate, _ := ioutil.ReadFile(opt.ConfigPath + "/templates/default/_index_srclist.html")
	srclist := string(srclistTempate)
	exp, _ := regexp.Compile("@SRCPATH@")
	srclist = exp.ReplaceAllString(srclist, filename)
	exp, _ = regexp.Compile("@VIOLATIONS@")
	srclist = exp.ReplaceAllString(srclist, fmt.Sprintf("%d", len(vs)))

	f.WriteString(srclist + newlineDefault)
	f.Close()

	var rootPath = "../"
	for c := 0; c < strings.Count(filename, "/"); c++ {
		rootPath = rootPath + "../"
	}
	var indexPath = rootPath + "index.html"

	fileexp, _ := regexp.Compile("/[^/]*$")
	var dirname = fileexp.ReplaceAllString(filename, "")
	os.MkdirAll(opt.Html+"/src/"+dirname, 0777)
	os.MkdirAll(opt.Html+"/js", 0777)
	os.MkdirAll(opt.Html+"/css", 0777)
	CopyFile(opt.ConfigPath+"/templates/default/js/src.js", opt.Html+"/js/src.js")
	CopyFile(opt.ConfigPath+"/templates/default/css/main.css", opt.Html+"/css/main.css")
	CopyFile(opt.ConfigPath+"/templates/default/css/index.css", opt.Html+"/css/index.css")
	CopyFile(opt.ConfigPath+"/templates/default/css/src.css", opt.Html+"/css/src.css")

	fdetail, _ := os.OpenFile(opt.Html+"/src/"+filename+".html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fdetail.Close()
	fdetail.WriteString(`<!DOCTYPE html>
<html><head>
<title>fint result</title>
<link rel="stylesheet" type="text/css" href="`)
	fdetail.WriteString(rootPath + "css/main.css")
	fdetail.WriteString(`" />
<link rel="stylesheet" type="text/css" href="`)
	fdetail.WriteString(rootPath + "css/src.css")
	fdetail.WriteString(`" />
<script type="text/javascript" src="`)
	fdetail.WriteString(rootPath + "js/src.js")
	fdetail.WriteString(`"></script>
</head><body>
<h1><a href="`)
	fdetail.WriteString(indexPath)
	fdetail.WriteString(`">fint result</a></h1>

<h2>Source File</h2>
<p>`)
	fdetail.WriteString(filename)
	fdetail.WriteString(`</p>

<h2>Result</h2>
<div id="srclist">
<table id="src">
<tbody>
`)

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
		fdetail.WriteString(fmt.Sprintf("<tr class=\"%s\" ", vsclass))
		if 0 < len(vs) {
			fdetail.WriteString(fmt.Sprintf("onclick=\"toggleMessages('msg_l%d');\"", n))
		}
		fdetail.WriteString(fmt.Sprintf("><td class=\"line\">%d</td><td class=\"code\"><pre>%s</pre></td></tr>\r\n", n, line))
		if 0 < len(vs) {
			fdetail.WriteString(fmt.Sprintf("<tr id=\"msg_l%d\" class=\"row_msg\"><td class=\"line\">&nbsp;</td><td class=\"list\">", n))
			for i := range vs {
				fdetail.WriteString(fmt.Sprintf("<div class=\"msg\">%s</div>", vs[i].Message))
			}
			fdetail.WriteString("</td></tr>\r\n")
		}
		if err == io.EOF {
			break
		}
	}
	fsrc.Close()

	fdetail.WriteString(`</tbody>
</table>
</div>
</body>
</html>
`)
	fdetail.Close()
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
	srclistTemp, _ := ioutil.ReadFile(opt.Html + "/_index_srclist.html")
	replaceTagInFile(opt.Html+"/index.html", "@SRCLIST@", string(srclistTemp))
	os.Remove(opt.Html + "/_index_srclist.html")
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
	conf, err = ioutil.ReadFile(opt.ConfigPath + "/config.json")
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
