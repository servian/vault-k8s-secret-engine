.PHONY: build clean kube-up kube-down run-vault

build:
	go build -o vault/plugins/vault-k8s-secret-engine cmd/main.go

clean:
	rm -rf ./vault

kube-up:
	kind create cluster
	cd tests && ./local_configure.sh

kube-down:
	kind delete cluster

kube-configure:
	cd tests && .local_configure.sh

run-vault:
	cd tests && ./local_build.sh