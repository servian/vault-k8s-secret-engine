# Vault K8s Dynamic Service Accounts
This project contains the source code for a [Hashicorp Vault](https://www.vaultproject.io/) plugin that provides on-demand (dynamic) credentials for a
short-lived [k8s](https://kubernetes.io/) service account.

This keeps the blast radius relatively small in case the credentials get leaked or abused.
