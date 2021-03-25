# Plugin local testing approach

## Set up local k8s cluster


## Local development
1. Pre-requisites
   - [Install Golang](https://golang.org/doc/install)
   - [Install Hashicorp Vault](https://learn.hashicorp.com/tutorials/vault/getting-started-install?in=vault/getting-started)
   - Kubernetes cluster to test against 
      - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
      - [minikube](https://minikube.sigs.k8s.io/docs/start/)
1. `./local_build.sh`
1. `./local_test.sh`