package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathK8sServiceAccount(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "k8s/service_account/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"role_name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Name of the role to associate with the service account",
			},
			"ttl": &framework.FieldSchema{
				Type:        framework.TypeDurationSecond,
				Description: "Lifetime of the returned service account credentials",
				Default:     300,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleReadOrCreate,
				Summary:  "Retrieve service account credentials",
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.handleReadOrCreate,
				Summary:  "Create service account credentials",
			},
		},
	}
}

func (b *backend) handleReadOrCreate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println(d.Raw)
	return b.secretAccessKeysCreate(ctx, req.Storage, "some string passed from handleReadOrCreate")
}
