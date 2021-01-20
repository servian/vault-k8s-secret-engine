export VAULT_ADDR="http://127.0.0.1:8200" \
&& vault login root \
&& vault secrets enable -path=k8s-dynamic-service-accounts vault-plugin-k8s-dynamic-service-accounts \
&& vault secrets list \
&& vault write k8s-dynamic-service-accounts/foo message="Hello World" \
&& vault read k8s-dynamic-service-accounts/foo
