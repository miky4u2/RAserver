#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o ./runtime/bin/server  ./server/server.go