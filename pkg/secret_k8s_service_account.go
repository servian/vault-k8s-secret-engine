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

func (b *backend) secretAccessKeysCreate(ctx context.Context, s logical.Storage, roleName string, kubeConfigPath string, ttl int) (*logical.Response, error) {
	b.Logger().Info(fmt.Sprintf("creating secret for role: %s with ttl: %d via kubeconfig at: %s", roleName, ttl, kubeConfigPath))
	resp := b.Secret(secretAccessKeyType).Response(map[string]interface{}{
		"ca.crt":    roleName,
		"namespace": "some namespace",
		"token":     "some token",
	}, map[string]interface{}{
		"some_internal_field": "some internal field value",
	})

	dur, err := time.ParseDuration(fmt.Sprintf("%ds", ttl))
	if err != nil {
		return nil, fmt.Errorf("error: %s occured when generating lease duration from ttl: %d", err, ttl)
	}
	resp.Secret.MaxTTL = dur
	resp.Secret.Renewable = false

	return resp, nil
}

func (b *backend) renewSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("renewing secret")
	return nil, fmt.Errorf("intentionally failing renewal of secret")
}

func (b *backend) revokeSecret(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.Logger().Info("revoking secret")
	return nil, fmt.Errorf("intentionally failing revokal of secret")
}
