package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const ttlInSeconds = 10
const key = "role_name"

func pathK8sServiceAccount(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account",
		Fields: map[string]*framework.FieldSchema{
			key: &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the role to associate with the service account",
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
		roleName := d.Get(key).(string)
		b.Logger().Info(fmt.Sprintf("role name: %s", roleName))
		return b.secretAccessKeysCreate(ctx, req.Storage, roleName, ttlInSeconds)
	} else {
		return nil, fmt.Errorf("could not find a role name to associated with the service account")
	}
}
