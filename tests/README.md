# Local development approach

## pre-requisites

- Golang >= 1.16
- Kind
- jq
- make

## Set up local k8s cluster

To test the plugin, you need to set up and configure a Kubernetes cluster for Vault to integrate with. Kind is the easiest way to do this, as it deploys a k8s cluster inside a docker contianer, making spin up and down super quick and easy.

Find the instalation instruction here: [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

Once that is set up, use the make file in the root of the repositiry to start and configure kuberenetes

**Note:** make sure the vault cluster has started before you run the make command

```sh
make kube-up
```

To tear down the cluster, call the matching down command

```sh
make kube-down
```

## Running the plugin

To run the plugin, it needs to be compiled, placed in the right location, and then vault needs to be started. To simplify that, make has a command for this.

**Note:** this is a sessino blocking command, so it's recommended to do this in a separate terminal window to the one you are working in

```sh
make vault-run
```

to stop vault, hit `ctrl+c`

## Configuring the integration between kube and vault

Once kube is running and vault is running, you need to create a service account in kube that has access to create new service accounts and set up bindings. Theres are a set of manifest files in the test folder under `tests/k8_manifests` that will create these in kube. They will also set up an example ClusterRole to test the plugin with called `reader_role`

The details of this service account is then used to configure the plugin

```sh
make kube-configure
```
## Testing

With everything set up, you can now test the plugin and confirm that the expected outcomes are achieved

**example test commnd:**
```sh
vault read k8s/service_account/default/viewer ttl_seconds=30 
```