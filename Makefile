.PHONY: build clean kube-up kube-down run-vault

build:
	go build -o vault/plugins/vault-k8s-secret-engine cmd/main.go

clean:
	rm -rf ./vault

kube-up:
	kind create cluster

kube-down:
	kind delete cluster

vault-configure:
	cd tests && ./local_configure.sh

vault-run:
	cd tests && ./local_build.sh