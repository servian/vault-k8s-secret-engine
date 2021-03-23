#!/bin/sh

kubectl apply -f tests/k8_manifests
export secret_name=$(kubectl get serviceaccount/vault-dynamic-creds-backend -o jsonpath='{.secrets[0].name}')
export k8_cacert=$(kubectl get secret/${secret_name} -o jsonpath='{.data.ca\.crt}'|base64 --decode)
export sa_token=$(kubectl get secret/${secret_name} -o jsonpath='{.data.token}'|base64 -d)

echo "token: ${sa_token}"
echo "cacert: ${k8_cacert}"
echo "---"
echo "K8 CACert and sa token set to env vars 'k8_cacert' and 'sa_token'"