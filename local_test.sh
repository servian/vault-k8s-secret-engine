export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s dsa \
&& vault secrets list \
&& vault write k8s/service_account role_name="my_super_role" kube_config_path="asdasd asdasd/asdasd asdasd" ttl_seconds="10" \
&& vault read k8s/service_account
