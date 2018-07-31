#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH=$GOPATH":$CURDIR"

OLDGOBIN="$GOBIN"

GOBIN="$CURDIR/bin/"
#gofmt -w src

go install logClient
go install logServer

export GOPATH="$OLDGOPATH"
export GOBIN="$OLDGOBIN"
echo 'install ok'