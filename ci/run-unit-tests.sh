#!/bin/bash

set -xe

source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  go test -v -race $(glide nv)
popd
