export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s dsa \
&& vault secrets list \
&& vault write k8s/service_account kube_config_path="/home/jigar/.kube/config" ttl_seconds="10" namespace="default"
