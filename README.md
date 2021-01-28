# Vault K8s Dynamic Service Accounts

This project contains the source code for a [Hashicorp Vault](https://www.vaultproject.io/) plugin that provides
on-demand (dynamic) credentials for a short-lived [k8s](https://kubernetes.io/) service account.

This keeps the blast radius relatively small in case the credentials get leaked or abused.

## How does it work?

1. Vault user requests credentials for a given k8s role
1. Vault plugin creates a service account for that role in k8s
1. Vault plugin retrieves the service account credentials and saves them in vault, with a ttl (configured with the
   plugin)
1. Vault plugin responds to the user request with credentials from step 3
1. Vault plugin lifecycle ensures that the k8s service account is deleted after the ttl from step 3 elapses

## Build

```
go build -o vault/plugins/dsa cmd/main.go
```

## Local development
1. Pre-requisites
   - [Install Golang](https://golang.org/doc/install)
   - [Install Hashicorp Vault](https://learn.hashicorp.com/tutorials/vault/getting-started-install?in=vault/getting-started)
   - Spin up a kubernetes cluster (locally or on the cloud)
      - place a kube config file for the cluster in a local folder
      - use the fully qualified path of the kube config file as the value for the parameter `kube_config_path` in `local_test.sh`
1. `./local_build.sh`
1. `./local_test.sh`
