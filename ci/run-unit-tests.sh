#!/bin/bash

set -xe

source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  echo "PWD: $(pwd)"
  echo "GOPATH: $GOPATH"
  go version
  glide -v
  glide install
  go test -v -race $(glide nv)
popd
