#!/bin/bash

set -xe

stage_dir="$PWD/stage"
source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  mkdir -p $stage_dir/release
  GOOS=linux GOARCH=amd64 go build -o $stage_dir/release/magnet github.com/pivotalservices/magnet/cmd/magnet
  git describe --abbrev=0 --tags > $stage_dir/tag
popd
