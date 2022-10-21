#!/bin/sh
pgrep webserver && kill -2 "$(pgrep webserver)"
setsid -f go run ./webserver.go >/dev/null 2>&1
