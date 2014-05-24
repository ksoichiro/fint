# fint

[![Build Status](https://travis-ci.org/ksoichiro/fint.svg?branch=master)](https://travis-ci.org/ksoichiro/fint)
[![Coverage Status](https://coveralls.io/repos/ksoichiro/fint/badge.png?branch=master)](https://coveralls.io/r/ksoichiro/fint?branch=master)

**fint** is a lightweight, simple source code check tool,  
but doesn't have syntax analysis feature -- so it is a fake lint :P  
**fint** is portable, executable on multiple platform, easy to integrate into your build process.

![Example](docs/screenshot.png)

## Installation

[Get the latest release binary](https://github.com/ksoichiro/fint/releases/latest) for your environment.

You can also install it from master branch with golang environment.  
This is slightly unstable than release binaries, but may have some new useful features.

```sh
$ go get github.com/ksoichiro/fint
```

## Usage

### For Objective-C, execute as a Xcode's "Run Script" of build phase

If your `GOPATH` is `~/go`, then put the following command
to shell script form:

```sh
TERM=dumb ~/go/bin/fint -s ${SRCROOT}/${PROJECT_NAME} -i objc
```

If format error found, the command will exit with code 1, otherwise 0.  
The results will be shown to the source code like normal syntax warnings.

See sample Xcode projects in [`testdata/objc`](testdata/objc) directory for details.

### Execute on command line

If you exported `GOPATH` and `${GOPATH}/bin` is in your `PATH` then:

```sh
$ fint -s testdata/objc/FintExample -i objc
testdata/objc/FintExample/FintExample/FEAppDelegate.m:12:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:14:1: warning: Space must be inserted between ']' and following message
testdata/objc/FintExample/FintExample/FEAppDelegate.m:14:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:18:1: warning: Space must be inserted before if
testdata/objc/FintExample/FintExample/FEAppDelegate.m:18:1: warning: Space must be inserted between ')' and '{'
testdata/objc/FintExample/FintExample/FEAppDelegate.m:19:1: warning: Space must be inserted between '//' and following comment
testdata/objc/FintExample/FintExample/FEAppDelegate.m:20:1: warning: Space must be inserted before else
testdata/objc/FintExample/FintExample/FEAppDelegate.m:20:1: warning: Space must be inserted before if
testdata/objc/FintExample/FintExample/FEAppDelegate.m:20:1: warning: Space must be inserted between ')' and '{'
testdata/objc/FintExample/FintExample/FEAppDelegate.m:22:1: warning: Space must be inserted after else
testdata/objc/FintExample/FintExample/FEAppDelegate.m:22:1: warning: Space must be inserted before else
testdata/objc/FintExample/FintExample/FEAppDelegate.m:24:1: warning: Space must be inserted after ','
testdata/objc/FintExample/FintExample/FEAppDelegate.m:32:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:33:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:38:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:39:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:44:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:49:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/FEAppDelegate.m:54:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExample/main.m:15:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExampleTests/FintExampleTests.m:19:1: warning: Line length exceeds 80 characters
testdata/objc/FintExample/FintExampleTests/FintExampleTests.m:24:1: warning: Line length exceeds 80 characters

22 warnings generated.
```

### Command line options

| Option | Description                                            |
| ------ | ------------------------------------------------------ |
| `-i`   | ID of the target rule sets. Required.                  |
| `-s`   | Source directory to be checked. Required.              |
| `-c`   | Config files directory. Default value is `.fint`.      |
| `-l`   | Message locale. Default value is `en`(English). Currently, `en` and `ja` is supported. |
| `-h`   | HTML report directory. Optional.                       |
| `-f`   | Force generating report to existing directory. Default is `false`. |
| `-q`   | Quiet mode. Suppresses output. Default is `false`.     |
| `-template` | HTML report template name. Default is `default`.  |

## Configuration

### Structure

    .fint // can be changed by `-c` option
    └── builtin
        ├── modules
        │   ├── max_length
        │   │   └── config.json
        │   └── pattern_match
        │       └── config.json
        ├── targets
        │   └── objc
        │       ├── locales
        │       │   ├── en.json
        │       │   └── ja.json
        │       └── ruleset.json
        └── templates
            └── default
                ├── _index.html
                ├── _index_srclist.html
                ├── _src.html
                ├── _src_srcline.html
                ├── _src_violation_msg.html
                ├── _src_violation_msglist.html
                ├── css
                │   ├── index.css
                │   ├── main.css
                │   └── src.css
                └── js
                    └── src.js

### Rule sets

Config file includes lint rule sets(`ruleset.json`), which is the top level element. Array.

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
| `rules` > `id`   | ID of the rule. |
| `rules` > `args` | Argument for the lint logic. Optional. |
| `rules` > `message` | Message to show when there is a violation of the rule. This is defined not in the `ruleset.json` but in the `locales/[LOCALE].json`. |

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
