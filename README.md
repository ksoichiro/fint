# fint

[![Build Status](https://travis-ci.org/ksoichiro/fint.svg?branch=master)](https://travis-ci.org/ksoichiro/fint)
[![Coverage Status](https://coveralls.io/repos/ksoichiro/fint/badge.png?branch=master)](https://coveralls.io/r/ksoichiro/fint?branch=master)

**fint** is a lightweight, simple source code check tool,  
but doesn't have syntax analysis feature -- so it is a fake lint :P  
**fint** is portable, executable on multiple platform, easy to integrate into your build process.

![Example](https://raw.githubusercontent.com/ksoichiro/fint/master/docs/screenshot.png)

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

See sample Xcode projects in [`testdata/objc`](https://github.com/ksoichiro/fint/tree/master/testdata/objc) directory for details.

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
| `-f`   | Force generating report to existing directory. Default value is `false`. |
| `-q`   | Quiet mode. Suppresses output. Default value is `false`. |
| `-template` | HTML report template name. Default value is `default`.  Currently, `default` and `dark` is available. |
| `-fix` | Fix violations if possible. Default is `false`. |

## Configuration

### Structure

    .fint
    └── builtin
        ├── modules
        │   ├── max_length
        │   │   └── config.json
        │   └── pattern_match
        │       └── config.json
        ├── targets
        │   ├── objc
        │   │   ├── locales
        │   │   │   ├── en.json
        │   │   │   └── ja.json
        │   │   └── ruleset.json
        │   └── sh
        │       ├── locales
        │       │   ├── en.json
        │       │   └── ja.json
        │       └── ruleset.json
        └── templates
            ├── dark
            │   ├── _index.html
            │   ├── _index_srclist.html
            │   ├── _src.html
            │   ├── _src_srcline.html
            │   ├── _src_violation_msg.html
            │   ├── _src_violation_msglist.html
            │   ├── css
            │   │   ├── index.css
            │   │   ├── main.css
            │   │   └── src.css
            │   └── js
            │       └── src.js
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

### Configuration root directory

All the configurations for `fint` must be included in the `.fint` directory.  
This directory can be changed by `-c` option.

### Targets

`fint` needs to know which modules to use for lint and how to use the modules.  
"Target" resolves them for a certain language or project.  
Targets are located in `.fint/targets`.  
To select a target, specify subdirectory name with `-i` option.  
Available targets:

* objc
* sh

In each target directories, `ruleset.json` must be located.  
This file defines the lint rule sets in the JSON format.

| Item  | Description |
| ----- | ----------- |
| `rulesets` | JSON array that includes the rule sets. Target can have multiple rule sets because the projects will have multiple file-types and they need multiple rules for lint. |
| `rulesets` > `id` | ID of the rule set. Currently, this is just a comment and not used for lint. |
| `rulesets` > `description` |  Description of this rule set. This will not be used from the program for now. |
| `rulesets` > `modules` |  Module configurations for this rule set. See 'Modules' for details. |

### Modules

"Module" means the lint logic which defines how to check source files.  
Modules' has several descriptions and configurations.

#### Description

Descriptions for each modules are located in `.fint/modules`.  
In `.fint/module/[MODULE_NAME]/config.json`, following information is defined:

| Item | Description |
| ---- | ----------- |
| `type` | Is the module built-in or external? |
| `executable` | If it's external module, where is it? |
| `description` | What is this module? |

#### Configuration

Each "targets" must have `ruleset.json` to define lint rule sets using "modules".  
So you should modify `ruleset.json` to configure modules.

A normal built-in module have the following configuration structure:

| Item | Description |
| ---- | ----------- |
| `id` | ID of the module. |
| `pattern` | File path pattern to select target source file. |
| `rules` | Rule for this modules. |
| `rules` > `id` | ID of the rule. This ID will be used in localization file. |
| `rules` > `args` | Arguments for the rule. Usage of this item will be different for each modules. |

#### Localization

Locale for lint warning messages.  
To select locales, specify locale name with `-l` option.  
Available locales:

* en
* ja

Localized messages are defined in `.fint/builtin/targets/[TARGET_NAME]/locales/[LOCALE].json`.

### HTML report

`fint` can output HTML report.  
To use this feature, specify reporting directory with `-h` option.

You can also configure HTML template.  
To select a template, specify subdirectroy name with `-template` option.  
Available templates:

* default
* dark

Templates are located in `.fint/builtin/templates/[TEMPLATE_NAME]`.

## Built-in modules

### Pattern match

This module checks if the line matching the pattern.  

| Item  | Description |
| ----- | ----------- |
| `id` | `pattern_match` |
| `rules` > `args` (0) | Forbidden pattern of the line. |
| `rules` > `args` (1) | Replacement string for auto-fix feature. |

### Max length

This module checks if the line exceeds a certain length.

| Item  | Description |
| ----- | ----------- |
| `id` | `max_length` |
| `rules` > `args` (0) | Pattern of the line to check length. |
| `rules` > `args` (1) | One element with max length. |

## License

Copyright (c) 2014 Soichiro Kashima  
Licensed under MIT license.  
See the bundled [LICENSE](https://github.com/ksoichiro/fint/blob/master/LICENSE) file for details.
