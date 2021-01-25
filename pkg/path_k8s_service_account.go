package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const maxTtlInSeconds = 10
const keyRoleName = "role_name"
const keyKubeConfigPath = "kube_config_path"
const keyTtlSeconds = "ttl_seconds"

func pathK8sServiceAccount(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account",
		Fields: map[string]*framework.FieldSchema{
			keyRoleName: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the role to associate with the service account",
			},
			keyKubeConfigPath: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Fully qualified path for the kubeconfig file to use",
			},
			keyTtlSeconds: &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: fmt.Sprintf("Time to live for the credentials returned. Must be <= %d seconds", maxTtlInSeconds),
				Default:     maxTtlInSeconds,
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
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleRead,
				Summary:  "Retrieve service account credentials",
			},
		},
	}
}

func (b *backend) handleRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return nil, fmt.Errorf("intentionally failing read")
}

// TODO: Check if we need to write to WAL in case of a replicated setup
func (b *backend) handleUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if d != nil {
		ttl := d.Get(keyTtlSeconds).(int)
		if ttl > maxTtlInSeconds {
			return nil, fmt.Errorf("%s cannot be more than %d", keyTtlSeconds, ttl)
		}

		roleName := d.Get(keyRoleName).(string)
		kubeConfigPath := d.Get(keyKubeConfigPath).(string)

		b.Logger().Info(fmt.Sprintf("role name: %s", roleName))
		return b.secretAccessKeysCreate(ctx, req.Storage, roleName, kubeConfigPath, ttl)
	} else {
		return nil, fmt.Errorf("could not find a role name to associated with the service account")
	}
}
