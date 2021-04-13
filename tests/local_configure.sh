#!/bin/sh

kubectl apply -f k8_manifests
export secret_name=$(kubectl get serviceaccount/vault-dynamic-creds-backend -o jsonpath='{.secrets[0].name}')
export k8_cacert=$(kubectl get secret/${secret_name} -o jsonpath='{.data.ca\.crt}'|base64 --decode)
export sa_token=$(kubectl get secret/${secret_name} -o jsonpath='{.data.token}'|base64 -d)
export server=$(kubectl config view --output json | jq -r '.clusters[] | select(.name=="kind-kind") | .cluster.server')

export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s vault-k8s-secret-engine \
&& vault secrets list \
&& vault write k8s/config viewer_role="reader_role" admin_role="admin_role" editor_role="editor_role" \
jwt="${sa_token}" \
ca_cert="${k8_cacert}" \
host="${server}" \
max_ttl=1h \
ttl=10m