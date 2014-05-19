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

Test data (source code) should be in `testdata` directory.

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
