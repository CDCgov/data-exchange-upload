#!/bin/sh

# Build the cosmos_sync.py Python script as a stand-alone binary.
# This is necesssary because the tusd default public image does not include
# any scripting languages like Python, NodeJS, nor Java run-time, etc.
# We don't want to build a custom image of tusd to include this since doing
# so would make tusd upgrades more difficult.
#
# cross-compile using the six8/pyinstaller-alpine linux docker image since
# the target platform for the public tusd docker image is linux
#
# output of this step will be dist/cosmos_sync
docker run --rm \
    -v "$PWD:$PWD" \
    -w "$PWD" \
    six8/pyinstaller-alpine:alpine-3.6-pyinstaller-v3.4 \
    --noconfirm \
    --onefile \
    --log-level DEBUG \
    --clean \
    cosmos_sync.py

# copy the output file and rename
cp dist/cosmos_sync cosmos-sync-bin
