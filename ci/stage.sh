#!/bin/bash

set -xe

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

stage_dir="$PWD/stage"
source_dir="$(cd "$(dirname "$0")" && pwd)"
pushd $source_dir/..
  glide install
  mkdir -p $stage_dir/release
  shorthash=$(git rev-parse --short HEAD)
  tag=$(git describe --abbrev=0 --tags)
  GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X main.Version=$tag-$shorthash" -o $stage_dir/release/magnet-linux ./cmd/magnet
  GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s -X main.Version=$tag-$shorthash" -o $stage_dir/release/magnet-darwin ./cmd/magnet
  GOOS=windows GOARCH=amd64 go build -ldflags "-w -s -X main.Version=$tag-$shorthash" -o $stage_dir/release/magnet.exe ./cmd/magnet
  echo $tag > $stage_dir/tag
popd
