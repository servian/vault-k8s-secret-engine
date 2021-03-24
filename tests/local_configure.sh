#!/bin/sh

kubectl apply -f k8_manifests
export secret_name=$(kubectl get serviceaccount/vault-dynamic-creds-backend -o jsonpath='{.secrets[0].name}')
export k8_cacert=$(kubectl get secret/${secret_name} -o jsonpath='{.data.ca\.crt}'|base64 --decode)
export sa_token=$(kubectl get secret/${secret_name} -o jsonpath='{.data.token}'|base64 -d)

export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s vault-k8s-secret-engine \
&& vault secrets list \
&& vault write k8s/config viewer_role="reader_role" \
jwt="${sa_token}" \
ca_cert="${k8_cacert}" \
base_url="https://127.0.0.1:64740"