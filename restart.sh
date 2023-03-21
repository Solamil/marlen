#!/bin/sh

export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:"$PATH"
web_dir="web/"
go test -coverprofile cover.out ./main.go ./main_test.go
go tool cover -html=cover.out -o ${web_dir}cover.html
printf "Closing old instance: %d" "$(pgrep main -u "$(whoami)")"
kill -2 "$(pgrep main -u "$(whoami)")" 2>&1
setsid -f go run ./main.go >/dev/null 2>&1
printf "Started new instance: %d" "$(pgrep main -u "$(whoami)")"
