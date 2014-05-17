// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const (
	errPrefix = "fint: "
	bufSize   = 4096
)

type Opt struct {
	SrcRoot    string
	ConfigPath string
	Locale     string
	Id         string
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
)

func newError(message string) error {
	return errors.New(errPrefix + message)
}

func getOpts() (err error) {
	srcRoot := flag.String("s", "", "Source directory")
	configPath := flag.String("c", "conf/config.json", "Config file path")
	locale := flag.String("l", "default", "Message locale")
	id := flag.String("i", "", "ID of the rule set")
	flag.Parse()
	if *srcRoot == "" {
		err = newError("source directory is required.")
		return
	}
	if *id == "" {
		err = newError("ID of the rule set is required.")
		return
	}
	opt = &Opt{SrcRoot: *srcRoot, ConfigPath: *configPath, Locale: *locale, Id: *id}
	return
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

func LoadConfig(file []byte) *Config {
	var c Config
	json.Unmarshal(file, &c)
	return &c
}

func checkSourceFile(filename string, rs RuleSet) (vs []Violation, err error) {
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		err = newError("cannot open " + filename)
		return
	}
	defer f.Close()
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

func checkFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	rs, errRs := findRuleSet()
	if errRs != nil {
		return errRs
	}

	if matched, _ := regexp.MatchString(rs.Pattern, path); matched {
		v, e := checkSourceFile(path, rs)
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

func Execute(o *Opt) (v []Violation, err error) {
	opt = o

	var conf []byte
	conf, err = ioutil.ReadFile(opt.ConfigPath)
	if err != nil {
		return
	}
	config = LoadConfig(conf)

	err = filepath.Walk(opt.SrcRoot, checkFile)
	return violations, err
}

func ExecuteAsCommand() {
	err := getOpts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	term = os.Getenv("TERM")

	_, err = Execute(opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := range violations {
		printViolation(violations[i])
	}

	if 0 < len(violations) {
		fmt.Printf("\n%d %s generated.\n",
			len(violations), pluralize(len(violations), "warning", "warnings"))
		os.Exit(1)
	}
}
