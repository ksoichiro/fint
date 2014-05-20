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
)

type Opt struct {
	SrcRoot    string
	ConfigPath string
	Locale     string
	Id         string
	Html       string
	Force      bool
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
	f, _ := os.OpenFile(opt.Html+"/index.html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString(`<!DOCTYPE html>
<html><head>
<title>fint result</title>
<style type="text/css">
body{background:#fff;}
table{border:1px solid #999;border-collapse:collapse;}
th,td{border:1px solid #999;padding:0.4em;}
</style>
</head><body>
<h1>fint result</h1>
<table>
<thead>
<tr><th>File</th><th>Violations</th></tr>
</thead>
<tbody>
`)

	f.Close()
}

func printReportBody(filename string, vs []Violation) {
	if opt.Html == "" {
		return
	}
	MkReportDir(true)
	var f *os.File
	f, _ = os.OpenFile(opt.Html+"/index.html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	var path = ".."
	for c := 0; c < strings.Count(strings.Trim(opt.Html, "/"), "/"); c++ {
		path = path + "../"
	}
	path = fmt.Sprintf("%s/%s", path, filename)
	f.WriteString(fmt.Sprintf("<tr><td><a href=\"%s\">%s</a></td><td>%d</td></tr>\r\n", path, path, len(vs)))

	f.Close()
}

func printReportFooter() {
	if opt.Html == "" {
		return
	}
	f, _ := os.OpenFile(opt.Html+"/index.html", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	f.WriteString(`</tbody>
</table>
</body></html>
`)

	f.Close()
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
		for i := range rs.Modules {
			switch rs.Modules[i].Id {
			case "pattern_match":
				for j := range rs.Modules[i].Rules {
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Pattern, line); matched {
						vs = append(vs, Violation{Filename: filename, Line: n, Message: rs.Modules[i].Rules[j].Message[opt.Locale]})
					}
				}
			case "max_length":
				for j := range rs.Modules[i].Rules {
					if matched, _ := regexp.MatchString(rs.Modules[i].Rules[j].Pattern, line); matched {
						max_len := int(rs.Modules[i].Rules[j].Args[0].(float64))
						if too_long := max_len < len(line); too_long {
							vs = append(vs, Violation{Filename: filename, Line: n, Message: fmt.Sprintf(rs.Modules[i].Rules[j].Message[opt.Locale], max_len)})
						}
					}
				}
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
	}

	printReportBody(filename, vs)

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
	conf, err = ioutil.ReadFile(opt.ConfigPath)
	if err != nil {
		return
	}
	config = LoadConfig(conf)

	printReportHeader()

	err = filepath.Walk(opt.SrcRoot, CheckFile)

	printReportFooter()

	return violations, err
}

func ExecuteAsCommand(o *Opt) (err error) {
	term = os.Getenv("TERM")
	_, err = Execute(o)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range violations {
		printViolation(violations[i])
	}

	if 0 < len(violations) {
		fmt.Printf("\n%d %s generated.\n",
			len(violations), pluralize(len(violations), "warning", "warnings"))
		err = newError("error while executing lint")
	}
	return
}
