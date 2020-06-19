#!/bin/bash
GOOS=darwin GOARCH=amd64 go build -o ./runtime/bin/server  ./server/server.go