#!/bin/sh

# Build the post_receive_bin.py Python script as a stand-alone binary.
# This is necesssary because the tusd default public image does not include
# any scripting languages like Python, NodeJS, nor Java run-time, etc.
# We don't want to build a custom image of tusd to include this since doing
# so would make tusd upgrades more difficult.
#
# cross-compile using the six8/pyinstaller-alpine linux docker image since
# the target platform for the public tusd docker image is linux
#
# output of this step will be dist/post-receive-bin
podman run --rm \
    -v "${PWD}:${PWD}" \
    -w "${PWD}" \
    ociodexdevupload.azurecr.io/pyinstaller:alpine-python-3.8.3 \
    --noconfirm \
    --onefile \
    --log-level DEBUG \
    --clean \
    post_receive_bin.py

# copy the output file and rename
cp target/post_receive_bin post-receive-bin
