# CONTRIBUTING

Just a memorandum for my work...

## Build

```sh
$ cd fint && go build && cd ..
```

This will generate `fint/fint`.

## Test

Execute test:

```sh
$ go test -coverprofile=profile.cov
```

Check where are not covered:

```sh
$ go tool cover -html=profile.cov
```

In short, execute this command:

```sh
$ go test -coverprofile=profile.cov && go tool cover -html=profile.cov
```

Test data (source code) should be in `testdata` directory.

### Auto-fix test

```sh
$ rm -rf testdata/objc/FintExampleFix
$ cp -pR testdata/objc/FintExample testdata/objc/FintExampleFix
$ go run fint/main.go -s testdata/objc/FintExampleFix -i objc -h report -f -fix
$ open report/index.html
```

## Test on Travis CI

If you push to this repository, build task on Travis CI will run.  
See `.travis.yml`.

## Cross-compile

Execute this:

```sh
$ ./crosscompile.sh
```

This will generate program binaries in `build` directory  
for each `GOOS`es and `GOARCH`s.
