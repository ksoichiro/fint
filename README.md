# fint - a fake of lint

[![Build Status](https://travis-ci.org/ksoichiro/fint.svg?branch=master)](https://travis-ci.org/ksoichiro/fint)
[![Coverage Status](https://coveralls.io/repos/ksoichiro/fint/badge.png?branch=master)](https://coveralls.io/r/ksoichiro/fint?branch=master)

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
| `-i`   | ID of the rule set.  Required.                         |
| `-s`   | Project source root directory. Required.               |
| `-c`   | Config file path. Default value is `.fint.json`. |
| `-l`   | Message locale. Default value is `default`(English). Currently, `default` and `ja` is supported. |

## Configuration

### Config file

Config file(JSON) is set by `-c` option.  

### Rule sets

Config file includes lint rule sets, which is the top level element. Array.

| Item  | Description |
| ----- | ----------- |
| `id` |  ID of the rule set. This will be used to select rule set by option `-i`. |
| `description` |  Description of this rule set. This will not be used from the program for now. |
| `pattern` |  File path pattern to apply this rule set. Regular expression. |
| `modules` |  Details of the rule set. See 'Modules'. |

### Modules

Modules describe parameters, warning messages for the lint logics.  
Each element of the array describes one lint logic type.  
Basic structure is below.

| Item  | Description |
| ----- | ----------- |
| `id` | ID of the lint logic. This is predefined in the program and not changeable. |
| `rules` | Array of the specific rules. |
| `rules` > `pattern` | Pattern for the lint logic. |
| `rules` > `args` | Argument for the lint logic. Optional. |
| `rules` > `message` | Message to show when there is a violation of the rule. It is an array which has default(`default`) and localized message(`ja`). |

Currently, the following modules are defined.

#### Pattern match

This module checks if the line matching the `pattern`.  

| Item  | Description |
| ----- | ----------- |
| `id` | `pattern_match` |
| `rules` > `pattern` | Forbidden pattern of the line. |
| `rules` > `args` | Not used. |

#### Max length

This module checks if the line exceeds a certain length.

| Item  | Description |
| ----- | ----------- |
| `id` | `max_length` |
| `rules` > `pattern` | Pattern of the line to check length. |
| `rules` > `args` | One element with max length. |

## License

Copyright (c) 2014 Soichiro Kashima  
Licensed under MIT license.  
See the bundled [LICENSE](LICENSE) file for details.
