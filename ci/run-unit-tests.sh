#!/bin/bash

set -xe

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  pwd
  echo $GOPATH
  go version
  glide -v
  glide install
  go test -v -race $(glide nv)
popd
