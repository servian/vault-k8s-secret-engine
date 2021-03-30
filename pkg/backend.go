package servian

import (
	"context"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// TODO: Finish help text
const backendHelp = `
The Vault dynamic service account backend provides on-demand, dynamic 
credentials for a short-lived k8s service account
`

// K8ServiceAccountFactory inits a new instance of the plugin
func K8sServiceAccountFactory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	k := KubernetesService{}
	b := Backend(&k)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// Backend instantiates the backend for the plugin
func Backend(k KubernetesInterface) *backend {
	var b backend
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		Paths: []*framework.Path{
			configurePlugin(&b),
			invalidPath(&b),
			readSecret(&b),
		},
		Secrets: []*framework.Secret{
			secret(&b),
		},
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				configPath,
			},
		},

		BackendType: logical.TypeLogical,
	}
	b.kubernetesService = k
	return &b
}

type backend struct {
	*framework.Backend
	kubernetesService KubernetesInterface
}
