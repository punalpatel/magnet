# magnet

[![Build Status](https://concourse.pivotalservices.io/api/v1/teams/magnet/pipelines/magnet/jobs/unit/badge)](https://concourse.pivotalservices.io/teams/magnet/pipelines/magnet)

## Installation

Dependencies are vendored with [Glide](https://github.com/Masterminds/glide)

Get the code with:

`$ go get -d -u github.com/pivotalservices/magnet`

Install dependencies:

```
$ cd $GOPATH/src/github.com/pivotalservices/magnet
$ glide install
```

Build/install the tool:

`$ go install`

## Environment Variables

`magnet` is configured via the following environment variables

```
export VSPHERE_SCHEME="https"                 # defaults to https
export VSPHERE_PORT="1234"                    # defaults to 443
export VSPHERE_INSECURE="true"                # Ignore SSL Cert Errors (defaults to false)
export VSPHERE_HOSTNAME="localhost"
export VSPHERE_USERNAME="administrator"
export VSPHERE_PASSWORD="password"
export VSPHERE_CLUSER="Cluster"
export VSPHERE_RESOURCEPOOL="RP01"           # optional
```
