# fint - a fake of lint

fint is a lightweight, simple source code check tool,  
but doesn't have syntax analysis feature -- so it is a fake lint.

fint is portable, executable on multiple platform, easy to integrate into your build process.

## Installation

For now, you can use this tool with golang environment.

    $ go get github.com/ksoichiro/fint

## Usage

### For Objective-C, execute as a Xcode's "Run Script" of build phase

If your `GOPATH` is `~/go`, then put the following command
to shell script form:

    ~/go/bin/fint -s ${SRCROOT} -p ${PROJECT_NAME}

If format error found, the command will exit with code 1, otherwise 0.

### Execute on command line

If you export `GOPATH` then:

    $ $GOPATH/bin/fint -s ~/Workspace/FormatCheck -p FormatCheck
    FormatCheck/FCAppDelegate.m:14:1: warning: format error

Or if you include `$GOPATH/bin` to `PATH` simply execute command:

    $ fint -s ~/Workspace/FormatCheck -p FormatCheck
    FormatCheck/FCAppDelegate.m:14:1: warning: format error

## License

Copyright (c) 2014 Soichiro Kashima  
Licensed under MIT license.  
See the bundled [LICENSE](LICENSE) file for details.
