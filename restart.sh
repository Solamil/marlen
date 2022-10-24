#!/bin/sh
pgrep main && kill -2 "$(pgrep main)"
setsid -f go run ./main.go >/dev/null 2>&1
