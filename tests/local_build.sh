#!/bin/sh

go build -o ../vault/plugins/vault-k8s-secret-engine ../cmd/main.go \
&& vault server -dev -dev-root-token-id=root -dev-plugin-dir=../vault/plugins
