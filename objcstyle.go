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

func grep(pattern, filename string) bool {
	var matchedAny bool = false
	f, err := os.Open(filename)
	if err != nil {
		return true
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, BufSize)
	for n := 1; true; n++ {
		lineBytes, isPrefix, err := r.ReadLine()
		if isPrefix {
			fmt.Printf("Too long line: %s", filename)
			return true
		}
		line := string(lineBytes)
		if err != io.EOF && err != nil {
			fmt.Println(err)
			return true
		}
		if matched, _ := regexp.MatchString(pattern, line); matched {
			matchedAny = true
			fmt.Printf("%s:%d:1: warning: format error\n", filename, n)
		}
		if err == io.EOF {
			break
		}
	}
	return matchedAny
}

func main() {
	opt, err := getOpts()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	os.Chdir(opt.SrcRoot)
	globPattern := opt.ProjName + "/*.[m|mm|h]"
	files, _ := filepath.Glob(globPattern)
	var matched bool = false
	for value := range files {
		filename := files[value]
		matched = matched || grep("else{", filename)
	}
	if matched {
		os.Exit(1)
	}
}
