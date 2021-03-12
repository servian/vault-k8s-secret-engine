export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s dsa \
&& vault secrets list \
&& vault write k8s/config max_ttl=50 allowed_roles="r1,r2" allowed_cluster_roles="r1c,r2c" \
kube_config="contents of kubeconfig file" \
&& vault read k8s/config \
&& vault read k8s/service_account namespace="my-awesome-ns" role_name="r1"
