// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package fint_test

import (
	"github.com/ksoichiro/fint"
	"testing"
)

func TestExecute(t *testing.T) {
	opt := &fint.Opt{SrcRoot: "testdata/objc/FormatCheck", ConfigPath: "conf/config.json", Locale: "default", Id: "objc"}
	fint.Execute(opt)
}
