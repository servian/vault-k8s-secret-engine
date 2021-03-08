package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const keyRoleName = "role_name"
const keyClusterRoleName = "cluster_role_name"
const keyKubeConfigPath = "kube_config_path"
const keyTtlSeconds = "ttl_seconds"
const keyCACert = "ca_cert"
const keyNamespace = "namespace"
const keyServiceAccountToken = "service_account_token"
const keyServiceAccountUID = "service_account_uid"
const keyServiceAccountName = "service_account_name"
const keyRoleBindingName = "role_binding_name"

func pathK8sServiceAccountForRole(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account",
		Fields: map[string]*framework.FieldSchema{
			keyRoleName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the kubernetes role to associated with a dynamic service account.",
			},
			keyNamespace: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The namespace under which the service account should be created",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleReadForRole,
				Summary:  "Create new service account credentials",
			},
		},
	}
}

// TODO: Check if we need to write to WAL in case of a replicated setup
func (b *backend) handleReadForRole(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if d != nil {
		roleName := d.Get(keyRoleName).(string)
		namespace := d.Get(keyNamespace).(string)
		return b.createSecret(ctx, req.Storage, namespace, roleName, RoleTypeRole)
	} else {
		return nil, fmt.Errorf("could not find a role name to associate with the service account")
	}
}
