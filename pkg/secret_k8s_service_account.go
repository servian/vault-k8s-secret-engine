package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"time"
)

const secretAccessKeyType = "access_keys"

func secretK8sServiceAccount(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretAccessKeyType,
		Fields: map[string]*framework.FieldSchema{
			"ca.crt": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "CA Cert",
			},
			"namespace": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Namespace",
			},
			"token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Token",
			},
		},
		Renew:  b.renewSecret,
		Revoke: b.revokeSecret,
	}
}

func (b *backend) secretAccessKeysCreate(ctx context.Context, s logical.Storage, whatever string) (*logical.Response, error) {
	fmt.Printf("creating secret: %s", whatever)
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"ca.crt":    whatever,
		"namespace": "some namespace",
		"token":     "some token",
	}, map[string]interface{}{
		"some_internal_field": "some internal field value",
	})

	dur, err := time.ParseDuration("1m")
	if err != nil {
		return nil, fmt.Errorf("error: %s occured when generating lease duration", err)
	}
	resp.Secret.MaxTTL = dur
	resp.Secret.Renewable = false

	return resp, nil
}

func (b *backend) renewSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println("renewing secret")
	return nil, fmt.Errorf("intentionally failing renewal of secret")
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	fmt.Println("revoking secret")
	return nil, fmt.Errorf("intentionally failing revokal of secret")
}
