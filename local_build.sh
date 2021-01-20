go build -o vault/plugins/vault-plugin-k8s-dynamic-service-accounts cmd/main.go \
&& vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins
