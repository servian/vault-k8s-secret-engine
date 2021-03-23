#!/bin/sh

go build -o vault/plugins/dsa cmd/main.go \
&& vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins
