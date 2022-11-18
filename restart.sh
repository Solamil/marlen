#!/bin/sh
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:"$PATH"
pgrep main && kill -2 "$(pgrep main)"
setsid -f go run ./main.go >/dev/null 2>&1
