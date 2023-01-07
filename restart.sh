#!/bin/sh

export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:"$PATH"
go test -coverprofile cover.out ./main.go ./main_test.go
go tool cover -html=cover.out -o cover.html
kill -2 "$(pgrep main -u "$(whoami)")" 2>&1
setsid -f go run ./main.go >> errors.log 
