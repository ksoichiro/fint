// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package main

import (
	"flag"
	"github.com/ksoichiro/fint"
)

func main() {
	var (
		srcRoot    = flag.String("s", "", "Source directory")
		configPath = flag.String("c", "conf/config.json", "Config file path")
		locale     = flag.String("l", "default", "Message locale")
		id         = flag.String("i", "", "ID of the rule set")
	)
	flag.Parse()

	fint.ExecuteAsCommand(&fint.Opt{SrcRoot: *srcRoot, ConfigPath: *configPath, Locale: *locale, Id: *id})
}
