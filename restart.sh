#!/bin/sh
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:"$PATH"
kill -2 "$(pgrep main -u "$(whoami)")" 2>&1
setsid -f go run ./main.go >/dev/null 2>&1
