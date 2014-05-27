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

type LintWalkFunc func(m common.Module, n int, filename, line, locale string) (vs []common.Violation)

func LintWalk(srcRoot string, m common.Module, locale string, lintWalkFunc LintWalkFunc) (fmap map[string]map[int][]common.Violation, err error) {
	if fmap == nil {
		fmap = make(map[string]map[int][]common.Violation)
	}

	fis, _ := ioutil.ReadDir(srcRoot)
	for i := range fis {
		entry := fis[i]
		filename := filepath.Join(srcRoot, entry.Name())
		if entry.IsDir() {
			var fmapSub map[string]map[int][]common.Violation
			fmapSub, err = LintWalk(filename, m, locale, lintWalkFunc)
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
			vsr := lintWalkFunc(m, n, filename, line, locale)
			if vsr != nil {
				lvs = append(lvs, vsr...)
			}
			vmap[n] = lvs
			if err == io.EOF {
				err = nil
				break
			}
		}
		if fmap[filename] == nil {
			fmap[filename] = make(map[int][]common.Violation)
		}
		fmap[filename] = vmap
	}
	return
}
