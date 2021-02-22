export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s dsa \
&& vault secrets list \
&& vault write k8s/config max_ttl=50 token_reviewer_jwt="some jwt token" \
kubernetes_host="http://localhost:5555" kubernetes_ca_cert="some cert data" \
allowed_roles="r1,r2" allowed_cluster_roles="r1c,r2c" \
&& vault read k8s/config
