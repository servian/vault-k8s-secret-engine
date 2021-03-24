package servian

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func invalidPath(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "service_account",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleInvalidPath,
				Summary:  "Create new service account credentials",
			},
		},
	}
}

// gives the user a nicer error message than the normal "No value found at .../service_account"
func (b *backend) handleInvalidPath(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return nil, fmt.Errorf("Invalid path, make sure the service account type is appended to the path, e.g. 'service_account/viewer'")
}
