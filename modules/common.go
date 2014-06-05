// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package modules

import (
	"bufio"
	"github.com/ksoichiro/fint/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LintWalkFunc func(m common.Module, n int, filename, line, locale string, shouldFix bool) (vs []common.Violation, fixedAny bool, fixedLine string)

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
			ftmp, _ = os.OpenFile(filename+".tmp", os.O_RDWR|os.O_CREATE, 0666)
			defer ftmp.Close()
		}
		if common.BufSize == 0 {
			common.BufSize = common.DefaultBufSize
		}
		r := bufio.NewReaderSize(f, common.BufSize)
		vmap := make(map[int][]common.Violation)
		for n := 1; true; n++ {
			var line string
			line, err = r.ReadString(common.LinefeedRune)
			if err != io.EOF && err != nil {
				return
			}
			var lvs []common.Violation
			vsr, fixedAny, fixedLine := lintWalkFunc(m, n, filename, strings.TrimSuffix(line, common.Linefeed), locale, fix)
			tmpLine := line
			if vsr != nil {
				lvs = append(lvs, vsr...)
				if fixedAny {
					if strings.HasSuffix(line, common.Linefeed) {
						tmpLine = fixedLine + common.Linefeed
					} else {
						tmpLine = fixedLine
					}
				}
			}
			if fix {
				ftmp.WriteString(tmpLine)
			}
			vmap[n] = lvs
			if err == io.EOF {
				err = nil
				break
			}
		}
		if fix {
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
	os.Remove(dst)
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
