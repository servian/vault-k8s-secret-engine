package servian

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"source.servian.com/eit-integration/common/vault-k8s-dynamic-service-accounts/pkg/k8s"
	"strings"
)

// TODO: Finish help text
const backendHelp = `
The Vault dynamic service account backend provides on-demand, dynamic 
credentials for a short-lived k8s service account
`
const maxTtlInSeconds = 300

func K8sServiceAccountFactory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	k := k8s.KubernetesService{}
	b := Backend(&k, maxTtlInSeconds)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	for k, v := range conf.Config {
		b.Logger().Info(fmt.Sprintf("Config: %s -> %s\n", k, v))
	}
	return b, nil
}

// TODO: implement a backend InitializeFunc to ensure we can connect to k8s
func Backend(k k8s.KubernetesInterface, maxTTLInSeconds int) *backend {
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
	b.kubernetesService = k
	b.maxTTLInSeconds = maxTTLInSeconds
	return &b
}

type backend struct {
	*framework.Backend
	maxTTLInSeconds   int
	kubernetesService k8s.KubernetesInterface
}
