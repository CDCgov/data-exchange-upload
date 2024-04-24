#!/bin/sh

# Go build
podman run --rm \
    -v "${PWD}:${PWD}" \
    -w "${PWD}" \
    golang:1.20.4-alpine \
    go build -o dist/post_finish_bin .

# copy the output file and rename
cp dist/post_finish_bin post-finish-bin
