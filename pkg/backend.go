package servian

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"strings"
)

// TODO: Finish help text
const backendHelp = `
The Vault dynamic service account backend provides on-demand, dynamic 
credentials for a short-lived k8s service account
`

func K8sServiceAccountFactory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func Backend() *backend {
	var b backend
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		Paths: []*framework.Path{
			pathK8sServiceAccount(&b),
		},
		Secrets: []*framework.Secret{
			secretK8sServiceAccount(&b),
		},
		BackendType: logical.TypeLogical,
	}

	return &b
}

type backend struct {
	*framework.Backend
}
