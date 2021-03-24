export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault read k8s/config \
&& vault read k8s/service_account/default/viewer_role
