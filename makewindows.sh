#!/bin/bash
#GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui -o ./runtime/bin/server.exe  /server/server.go
GOOS=windows GOARCH=amd64 go build -o ./runtime/bin/server.exe  ./server/server.go
