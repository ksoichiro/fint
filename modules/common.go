// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"bufio"
	"fmt"
	"github.com/ksoichiro/fint/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type LintWalkFunc func(m common.Module, n int, filename, line, locale string) (vs []common.Violation, fixedAny bool)

func LintWalk(srcRoot string, m common.Module, locale string, fix bool, lintWalkFunc LintWalkFunc) (fmap map[string]map[int][]common.Violation, err error) {
	if fmap == nil {
		fmap = make(map[string]map[int][]common.Violation)
	}

	fis, _ := ioutil.ReadDir(srcRoot)
	for i := range fis {
		entry := fis[i]
		filename := filepath.Join(srcRoot, entry.Name())
		if entry.IsDir() {
			var fmapSub map[string]map[int][]common.Violation
			fmapSub, err = LintWalk(filename, m, locale, fix, lintWalkFunc)
			if err != nil {
				return
			}
			// Merge into one map
			if fmapSub != nil {
				for f, vmap := range fmapSub {
					if fmap[f] == nil {
						fmap[f] = make(map[int][]common.Violation)
					}
					for n, vs := range vmap {
						fmap[f][n] = append(fmap[f][n], vs...)
					}
				}
			}
			continue
		}
		if matched, _ := regexp.MatchString(m.Pattern, filename); !matched {
			continue
		}
		var f *os.File
		f, err = os.Open(filename)
		if err != nil {
			err = common.NewError("cannot open " + filename)
			return
		}
		defer f.Close()
		var ftmp *os.File
		if fix {
			// Prepare fixed file
			CopyFile(filename, filename+".tmp")
			ftmp, _ = os.OpenFile(filename+".tmp", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			fmt.Printf("copied %s to %s\n", filename, filename+".tmp")
			defer ftmp.Close()
		}
		if common.BufSize == 0 {
			common.BufSize = common.DefaultBufSize
		}
		r := bufio.NewReaderSize(f, common.BufSize)
		vmap := make(map[int][]common.Violation)
		for n := 1; true; n++ {
			var (
				lineBytes []byte
				isPrefix  bool
			)
			lineBytes, isPrefix, err = r.ReadLine()
			if isPrefix {
				err = common.NewError(fmt.Sprintf("too long line: %s", filename))
				return
			}
			line := string(lineBytes)
			if err != io.EOF && err != nil {
				return
			}
			var lvs []common.Violation
			vsr, fixedAny := lintWalkFunc(m, n, filename, line, locale)
			tmpLine := line
			if vsr != nil {
				lvs = append(lvs, vsr...)
				if fixedAny {
					tmpLine = vsr[len(vsr)-1].Fix
				}
			}
			if fix {
				ftmp.WriteString(tmpLine + common.NewlineDefault)
			}
			vmap[n] = lvs
			if err == io.EOF {
				err = nil
				break
			}
		}
		if fix {
			fmt.Printf("closed %s\n", filename+".tmp")
			ftmp.Close()
			os.Remove(filename)
			CopyFile(filename+".tmp", filename)
			os.Remove(filename + ".tmp")
		}
		if fmap[filename] == nil {
			fmap[filename] = make(map[int][]common.Violation)
		}
		fmap[filename] = vmap
	}
	return
}

func CopyFile(src, dst string) (err error) {
	fin, err := os.Open(src)
	if err != nil {
		return
	}
	defer fin.Close()
	// Remove dst file once, if exists
	if _, err := os.Stat(dst); err != nil && os.IsExist(err) {
		os.Remove(dst)
	}
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
