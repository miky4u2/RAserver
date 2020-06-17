#!/bin/bash
GOOS=darwin GOARCH=amd64 go build -o ./runtime/bin/server.exe  ./server/server.go