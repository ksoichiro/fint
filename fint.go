/*
 Copyright (c) 2014 Soichiro Kashima
*/
package fint

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const bufSize = 4096

type Opt struct {
	SrcRoot    string
	ProjName   string
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

var opt *Opt
var config *Config

var violations []Violation
var term string

func getOpts() (*Opt, error) {
	srcRoot := flag.String("s", ".", "Project source root dir")
	projName := flag.String("p", "", "Project name")
	configPath := flag.String("c", "conf/config.json", "Config file path")
	locale := flag.String("l", "default", "Message locale")
	id := flag.String("i", "", "ID of the rule set")
	flag.Parse()
	opt := &Opt{SrcRoot: *srcRoot, ProjName: *projName, ConfigPath: *configPath, Locale: *locale, Id: *id}
	return opt, nil
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

func checkSourceFile(filename string, rs RuleSet) []Violation {
	var vs []Violation
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Cannot open " + filename)
		return vs
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, bufSize)
	for n := 1; true; n++ {
		lineBytes, isPrefix, err := r.ReadLine()
		if isPrefix {
			fmt.Printf("Too long line: %s", filename)
			return vs
		}
		line := string(lineBytes)
		if err != io.EOF && err != nil {
			fmt.Println(err)
			return vs
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
			break
		}
	}
	return vs
}

func findRuleSet() RuleSet {
	var rs RuleSet
	for i := range config.RuleSets {
		r := config.RuleSets[i]
		if r.Id == opt.Id {
			rs = r
		}
	}
	if rs.Id == "" {
		panic("No matching ruleset to [" + opt.Id + "]")
	}
	return rs
}

func checkFile(path string, f os.FileInfo, err error) error {
	rs := findRuleSet()
	if matched, _ := regexp.MatchString(rs.Pattern, path); matched {
		violations = append(violations, checkSourceFile(path, rs)...)
	}
	return nil
}

func pluralize(value int, singular, plural string) string {
	if value < 2 {
		return singular
	}
	return plural
}

func ExecuteAsCommand() {
	var err error
	opt, err = getOpts()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	term = os.Getenv("TERM")

	conf, err := ioutil.ReadFile(opt.ConfigPath)
	if err != nil {
		fmt.Println("Config file not found.")
		os.Exit(1)
	}
	config = LoadConfig(conf)

	os.Chdir(opt.SrcRoot)
	err = filepath.Walk(opt.ProjName, checkFile)
	for i := range violations {
		printViolation(violations[i])
	}

	if 0 < len(violations) {
		fmt.Printf("\n%d %s generated.\n",
			len(violations), pluralize(len(violations), "warning", "warnings"))
		os.Exit(1)
	}
}
