// Copyright (c) 2014 Soichiro Kashima
// Licensed under MIT license.

package main

import (
	"flag"
	"github.com/ksoichiro/fint"
	"github.com/ksoichiro/fint/common"
	"os"
)

const (
	ExitCodeSuccess = 0
	ExitCodeError   = 1
)

func main() {
	// Parse command
	if len(os.Args) == 1 {
		common.PrintUsage()
		os.Exit(ExitCodeError)
	}
	switch os.Args[1] {
	case "run":
	case "version":
		common.PrintVersion()
		os.Exit(ExitCodeSuccess)
	case "help":
		fallthrough
	default:
		common.PrintUsage()
		os.Exit(ExitCodeError)
	}

	// Parse flags
	var (
		srcRoot    = flag.String("s", "", "Source directory to be checked. Required.")
		configPath = flag.String("c", ".fint", "Config files directory. Default value is `.fint`.")
		locale     = flag.String("l", "en", "Message locale. Default value is `en`(English). Currently, `en` and `ja` is supported.")
		id         = flag.String("i", "", "ID of the target rule sets. Required.")
		html       = flag.String("h", "", "HTML report directory. Optional.")
		force      = flag.Bool("f", false, "Force generating report to existing directory. Default is `false`.")
		quiet      = flag.Bool("q", false, "Quiet mode. Suppresses output. Default is `false`.")
		template   = flag.String("template", "default", "HTML report template name. Default is `default`.")
		fix        = flag.Bool("fix", false, "Fix violations. Default is `false`.")
	)
	// Parse without filename and command
	flag.CommandLine.Parse(os.Args[2:])

	err := fint.ExecuteAsCommand(
		&common.Opt{
			SrcRoot:    *srcRoot,
			ConfigPath: *configPath,
			Locale:     *locale,
			Id:         *id,
			Html:       *html,
			Force:      *force,
			Quiet:      *quiet,
			Template:   *template,
			Fix:        *fix})
	if err != nil {
		os.Exit(ExitCodeError)
	}
}
