package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const keyRoleName = "role_name"
const keyKubeConfigPath = "kube_config_path"
const keyTtlSeconds = "ttl_seconds"
const keyCACert = "ca_cert"
const keyNamespace = "namespace"
const keyServiceAccountToken = "service_account_token"
const keyServiceAccountUID = "service_account_uid"
const keyServiceAccountName = "service_account_name"
const keyRoleBindingName = "role_binding_name"

func pathK8sServiceAccount(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account",
		Fields: map[string]*framework.FieldSchema{
			keyKubeConfigPath: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Fully qualified path for the kubeconfig file to use",
			},
			keyTtlSeconds: &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: fmt.Sprintf("Time to live for the credentials returned. Must be <= %d seconds", b.maxTTLInSeconds),
				Default:     b.maxTTLInSeconds,
			},
			keyNamespace: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The namespace under which the service account should be created",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.handleUpdate,
				Summary:  "Create new service account credentials",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.handleUpdate,
				Summary:  "Create new service account credentials",
			},
		},
	}
}

// TODO: Check if we need to write to WAL in case of a replicated setup
func (b *backend) handleUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if d != nil {
		ttl := d.Get(keyTtlSeconds).(int)
		if ttl > b.maxTTLInSeconds {
			return nil, fmt.Errorf("%s cannot be more than %d", keyTtlSeconds, b.maxTTLInSeconds)
		}

		kubeConfigPath := d.Get(keyKubeConfigPath).(string)
		namespace := d.Get(keyNamespace).(string)

		return b.createSecret(ctx, req.Storage, kubeConfigPath, ttl, namespace)
	} else {
		return nil, fmt.Errorf("could not find a role name to associate with the service account")
	}
}