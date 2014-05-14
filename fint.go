/*
 Copyright (c) 2014 Soichiro Kashima
*/
package main

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
}

type Rule struct {
	Pattern string
	Message map[string]string
}

type RuleSet struct {
	Id          string
	Description string
	Pattern     string
	Rules       []Rule
}

type FintConfig struct {
	RuleSets []RuleSet
}

var opt *Opt
var fintConfig *FintConfig
var violationCount int = 0

func getOpts() (*Opt, error) {
	srcRoot := flag.String("s", ".", "Project source root dir")
	projName := flag.String("p", "", "Project name")
	configPath := flag.String("c", "conf/config.json", "Config file path")
	locale := flag.String("l", "default", "Message locale")
	flag.Parse()
	opt := &Opt{SrcRoot: *srcRoot, ProjName: *projName, ConfigPath: *configPath, Locale: *locale}
	return opt, nil
}

func loadFintConfig(file []byte) *FintConfig {
	var fc FintConfig
	json.Unmarshal(file, &fc)
	return &fc
}

func checkSourceFile(filename string) int {
	var violationInFile int = 0
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Cannot open " + filename)
		return 1
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, bufSize)
	for n := 1; true; n++ {
		lineBytes, isPrefix, err := r.ReadLine()
		if isPrefix {
			fmt.Printf("Too long line: %s", filename)
			return 1
		}
		line := string(lineBytes)
		if err != io.EOF && err != nil {
			fmt.Println(err)
			return 1
		}
		var rs RuleSet = fintConfig.RuleSets[0]
		for i := range rs.Rules {
			if matched, _ := regexp.MatchString(rs.Rules[i].Pattern, line); matched {
				violationInFile++
				fmt.Printf("%s:%d:1: warning: %s\n", filename, n, rs.Rules[i].Message[opt.Locale])
			}
		}
		if err == io.EOF {
			break
		}
	}
	return violationInFile
}

func checkFile(path string, f os.FileInfo, err error) error {
	pattern := fintConfig.RuleSets[0].Pattern
	if matched, _ := regexp.MatchString(pattern, path); matched {
		violationCount += checkSourceFile(path)
	}
	return nil
}

func pluralize(value int, singular, plural string) string {
	if value < 2 {
		return singular
	}
	return plural
}

func main() {
	var err error
	opt, err = getOpts()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	conf, err := ioutil.ReadFile(opt.ConfigPath)
	if err != nil {
		fmt.Println("Config file not found.")
		os.Exit(1)
	}
	fintConfig = loadFintConfig(conf)

	os.Chdir(opt.SrcRoot)
	err = filepath.Walk(opt.ProjName, checkFile)

	if 0 < violationCount {
		fmt.Printf("\n%d %s generated.\n",
			violationCount, pluralize(violationCount, "warning", "warnings"))
		os.Exit(1)
	}
}
