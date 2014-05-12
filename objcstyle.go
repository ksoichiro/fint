/*
 Copyright (c) 2014 Soichiro Kashima
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const BufSize = 4096

type Opt struct {
	SrcRoot  string
	ProjName string
}

func getOpts() (*Opt, error) {
	srcRoot := flag.String("s", ".", "Project source root dir")
	projName := flag.String("p", "", "Project name")
	flag.Parse()
	opt := &Opt{SrcRoot: *srcRoot, ProjName: *projName}
	return opt, nil
}

func grep(pattern, filename string) int {
	var violationInFile int = 0
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Cannot open " + filename)
		return 1
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, BufSize)
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
		if matched, _ := regexp.MatchString(pattern, line); matched {
			violationInFile++
			fmt.Printf("%s:%d:1: warning: format error\n", filename, n)
		}
		if err == io.EOF {
			break
		}
	}
	return violationInFile
}

var violationCount int = 0

func checkFile(path string, f os.FileInfo, err error) error {
	if matched, _ := regexp.MatchString(".*\\.(m|mm|h)$", path); matched {
		violationCount += grep("else{", path)
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
	opt, err := getOpts()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	os.Chdir(opt.SrcRoot)
	err = filepath.Walk(opt.ProjName, checkFile)

	if 0 < violationCount {
		fmt.Printf("\n%d %s generated.\n",
			violationCount, pluralize(violationCount, "warning", "warnings"))
		os.Exit(1)
	}
}
