#!/bin/bash

TAG=
if [ $# -ge 1 ]; then
  TAG=$1
else
  TAG=`git rev-parse --short HEAD`
fi

function build() {
  echo "[1;35mBuilding for $1/$2...[m"
  local arch_dir=fint-$TAG-bin-$1-$2
  local bin_dir=build/${arch_dir}
  mkdir -p ${bin_dir}
  pushd fint > /dev/null 2>&1
  GOOS=$1 GOARCH=$2 go build
  local status=$?
  popd > /dev/null 2>&1
  if [ $status -eq 0 ]; then
    mv fint/fint* ${bin_dir}/
    cp -pR conf ${bin_dir}/
    pushd build > /dev/null 2>&1
    tar czf ${arch_dir}.tar.gz ${arch_dir}
    zip -ry ${arch_dir}.zip ${arch_dir} > /dev/null
    popd > /dev/null 2>&1
  fi
}

if [ -d build ]; then
  rm -rf build/*
fi

build darwin 386
build darwin amd64
build dragonfly 386
build dragonfly amd64
build freebsd 386
build freebsd amd64
build freebsd arm
build linux 386
build linux amd64
build linux arm
build netbsd 386
build netbsd amd64
build netbsd arm
build openbsd 386
build openbsd amd64
build windows 386
build windows amd64
