#!/bin/sh

if [ -e ".env" ]; then
    go run ./cmd/main.go -appconf=.env
else 
    go run ./cmd/main.go
fi
