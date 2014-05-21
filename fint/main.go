// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package main

import (
	"flag"
	"github.com/ksoichiro/fint"
	"os"
)

const (
	ExitCodeError = 1
)

func main() {
	var (
		srcRoot    = flag.String("s", "", "Source directory")
		configPath = flag.String("c", ".fint", "Config files directory")
		locale     = flag.String("l", "default", "Message locale")
		id         = flag.String("i", "", "ID of the rule set")
		html       = flag.String("h", "", "Generate result as HTML")
		force      = flag.Bool("f", false, "Force generating result to existing directory")
		quiet      = flag.Bool("q", false, "Quiet mode. Suppresses output.")
	)
	flag.Parse()

	err := fint.ExecuteAsCommand(
		&fint.Opt{
			SrcRoot:    *srcRoot,
			ConfigPath: *configPath,
			Locale:     *locale,
			Id:         *id,
			Html:       *html,
			Force:      *force,
			Quiet:      *quiet})
	if err != nil {
		os.Exit(ExitCodeError)
	}
}
