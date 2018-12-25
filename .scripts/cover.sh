#!/bin/bash -e
go test -covermode=atomic -cover -coverprofile cover.out -v -tags integration ./...

$HOME/gopath/bin/goveralls -coverprofile=cover.out -service travis-ci
rm -f cover.out