#!/bin/sh

export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s dsa \
&& vault secrets list \
&& vault write k8s/config max_ttl=50 allowed_roles="reader_role" \
jwt="${sa_token}" \
ca_cert="${k8_cacert}" \
base_url="https://127.0.0.1:64740"