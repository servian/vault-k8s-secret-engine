# Vault K8s Dynamic Service Accounts

This project contains the source code for a [Hashicorp Vault](https://www.vaultproject.io/) plugin that provides
on-demand (dynamic) credentials for a short-lived [k8s](https://kubernetes.io/) service account.

This keeps the blast radius relatively small in case the credentials get leaked or abused.

----

**Note:** This plugin is still under active development

----

## How does it work?
![overview](./docs/overview.png "Overview")

1. Vault user requests credentials for a given k8s role
1. Vault plugin creates a service account for that role in k8s
1. Vault plugin retrieves the service account credentials and saves them in vault, with a ttl (configured with the
   plugin)
1. Vault plugin responds to the user request with credentials from step 3
1. Vault plugin lifecycle ensures that the k8s service account is deleted after the ttl from step 3 elapses

## Quickstart

**ToDo:** This needs to be filled out

1. Download the plugin from the relase page (or download from source and build yourself)
1. Drop the plugin binary into the vault plugin directly and restart the vault service for it to pick up the new plugin
1. Configure a Service Account in the target kubernetes cluster with enough permissions to create Service Accounts and set up RoleBindings
1. Create a set of 3 ClusterRoles for the 3 types of Service Account roles this plugin can generate: admin, editor, viewer
1. Enble the secret engine by using the enable function of the vault cli `vault secrets enable -path=k8s vault-k8s-secret-engine`
1. Configure the secret engine
1. 

## Configuring the Secret engine



### Why ClusterRole instead of a Role object in Kubernetes?

A Role is scoped to a specific namespace, and cannot be used outside of that specific namespace. This means a map of role <-> namespace has to be created for each namespace in the cluster. And if a new namespace is added it will require a reconfiguration of the secrete backend. 

A ClusterRole on the other hand is scoped to the whole cluster. This means it can be re-used across all namespaces. We use a RoleBinding instead of a ClusterRolebinding to limit the scope of a binding to a single namespace.

This approach allows us to dynamically adjust to changes in the K8s cluster without compromising security. Vault policy is relied upon to limit access for a client to only the namespaces and service account type they require access to.

## Build

```
make build
```

## Local development
1. Pre-requisites
   - [Install Golang](https://golang.org/doc/install)
   - [Install Hashicorp Vault](https://learn.hashicorp.com/tutorials/vault/getting-started-install?in=vault/getting-started)
   - Kubernetes cluster to test against 
      - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
      - [minikube](https://minikube.sigs.k8s.io/docs/start/)
1. `./local_build.sh`
1. `./local_test.sh`

