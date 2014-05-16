# fint - a fake of lint

fint is a lightweight, simple source code check tool,  
but doesn't have syntax analysis feature -- so it is a fake lint.

fint is portable, executable on multiple platform, easy to integrate into your build process.

## Installation

For now, you can use this tool with golang environment.

```sh
$ go get github.com/ksoichiro/fint
```

## Usage

### For Objective-C, execute as a Xcode's "Run Script" of build phase

If your `GOPATH` is `~/go`, then put the following command
to shell script form:

```sh
$ ~/go/bin/fint -s ${SRCROOT}/${PROJECT_NAME} -i objc -l default
```

If format error found, the command will exit with code 1, otherwise 0.  
The results will be shown to the source code like normal syntax warnings.

### Execute on command line

If you export `GOPATH` then:

```sh
$ ${GOPATH}/bin/fint -s ~/Workspace/FormatCheck -p FormatCheck -i objc -l default
FormatCheck/FCAppDelegate.m:14:1: warning: format error
```

Or if you include `$GOPATH/bin` to `PATH` simply execute command:

```sh
$ fint -s ~/Workspace/FormatCheck -i objc -l default
FormatCheck/FCAppDelegate.m:14:1: warning: format error
```

### Command line options

| Option | Description                                            |
| ------ | ------------------------------------------------------ |
| `-c`   | Config file path. Default value is `conf/config.json`. |
| `-i`   | ID of the rule set.  Required.                         |
| `-l`   | Message locale. Default value is `default`(English). Currently, `default` and `ja` is supported. |
| `-s`   | Project source root directory. Default value is `.`.   |

## License

Copyright (c) 2014 Soichiro Kashima  
Licensed under MIT license.  
See the bundled [LICENSE](LICENSE) file for details.
