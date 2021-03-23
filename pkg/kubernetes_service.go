package servian

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-hclog"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const serviceAccountNamePrefix = "vault-sa-"
const roleNamePrefix = "vault-r-"
const roleBindingNamePrefix = "vault-rb-"

const serviceAccountKind = "ServiceAccount"
const roleKind = "Role"

type KubernetesService struct {
}

func (k *KubernetesService) CreateServiceAccount(pluginConfig PluginConfig, namespace string, l hclog.Logger) (*ServiceAccountDetails, error) {
	clientSet, err := getClientSet(pluginConfig, l)
	if err != nil {
		return nil, err
	}
	sa := v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: serviceAccountNamePrefix,
			Namespace:    namespace,
		},
	}
	sar, err := clientSet.CoreV1().ServiceAccounts(namespace).Create(&sa)
	if err != nil {
		return nil, err
	}

	var secrets []*string
	for _, item := range sar.Secrets {
		s := item.String()
		secrets = append(secrets, &s)
	}

	return &ServiceAccountDetails{
		Namespace: sar.Namespace,
		UID:       fmt.Sprintf("%s", sar.UID),
		Name:      sar.Name,
	}, nil
}

func (k *KubernetesService) GetServiceAccountSecret(pluginConfig PluginConfig, sa *ServiceAccountDetails, l hclog.Logger) ([]*ServiceAccountSecret, error) {
	clientSet, err := getClientSet(pluginConfig, l)
	if err != nil {
		return nil, err
	}

	ksa, err := clientSet.CoreV1().ServiceAccounts(sa.Namespace).Get(sa.Name, metav1.GetOptions{})

	var secrets []*ServiceAccountSecret
	for _, secret := range ksa.Secrets {
		secretName := secret.Name
		token, err := clientSet.CoreV1().Secrets(sa.Namespace).Get(secretName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if token != nil {
			caCert := string(token.Data["ca.crt"])
			tokenNamespace := string(token.Data["namespace"])
			tokenValue := string(token.Data["token"])
			secretValue := ServiceAccountSecret{
				CACert:    caCert,
				Namespace: tokenNamespace,
				Token:     tokenValue,
			}
			secrets = append(secrets, &secretValue)
		}
	}

	return secrets, nil
}

func (k *KubernetesService) DeleteServiceAccount(pluginConfig PluginConfig, namespace string, serviceAccountName string, l hclog.Logger) error {
	clientSet, err := getClientSet(pluginConfig, l)
	if err != nil {
		return err
	}
	err = clientSet.CoreV1().ServiceAccounts(namespace).Delete(serviceAccountName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubernetesService) CreateRoleBinding(pluginConfig PluginConfig, namespace string, serviceAccountName string, roleName string, l hclog.Logger) (*RoleBindingDetails, error) {
	clientSet, err := getClientSet(pluginConfig, l)
	if err != nil {
		return nil, err
	}
	subjects := []rbac.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}

	l.Info(fmt.Sprintf("subjects: %#v", subjects))

	roleBinding := rbac.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBindnig",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: roleBindingNamePrefix,
			Namespace:    namespace,
		},
		Subjects: subjects,
		RoleRef: rbac.RoleRef{
			Kind: "Role",
			Name: roleName,
		},
	}

	//l.Info(fmt.Sprintf("roleBinding: %#v", roleBinding))
	time.Sleep(2 * time.Second)

	rb, err := clientSet.RbacV1().RoleBindings(namespace).Create(&roleBinding)
	if err != nil {
		l.Info(fmt.Sprintf("rb-meta: %+v", roleBinding))
		l.Info(fmt.Sprintf("err: %+v", err))
		l.Info(fmt.Sprintf("rb: %+v", rb))
		return nil, err
	}
	return &RoleBindingDetails{
		Namespace: rb.Namespace,
		UID:       fmt.Sprintf("%s", rb.UID),
		Name:      rb.Name,
	}, nil
}

func (k *KubernetesService) DeleteRoleBinding(pluginConfig PluginConfig, namespace string, roleBindingName string, l hclog.Logger) error {
	clientSet, err := getClientSet(pluginConfig, l)
	if err != nil {
		return err
	}
	err = clientSet.RbacV1().RoleBindings(namespace).Delete(roleBindingName, nil)
	if err != nil {
		return err
	}
	return nil
}

func getClientSet(pluginConfig PluginConfig, l hclog.Logger) (*kubernetes.Clientset, error) {

	// using cleanhttp here rather than net/http so it's an isolated client and doesn't impact future instantiations.
	// Changes to http.DefaultClient persists for the lifetime of the process
	client := cleanhttp.DefaultClient()

	// Replace the trustsed cert pool with the one configured for the specific Kubernetes cluster
	// this will cause an SSL failure if the connection is directed to anything other than the expected cluster
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(pluginConfig.CACert))

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    certPool,
	}

	client.Transport.(*http.Transport).TLSClientConfig = tlsConfig

	clientConf := rest.ClientContentConfig{}
	clientConf.GroupVersion = v1.SchemeGroupVersion // I think this works???
	clientConf.Negotiator = runtime.NewClientNegotiator(scheme.Codecs.WithoutConversion(), clientConf.GroupVersion)

	//l.Info(fmt.Sprintf("Replaced certpool with cert: %s", pluginConfig.CACert))

	c, err := rest.NewRESTClient(pluginConfig.BaseUrl, "/api/v1", clientConf, nil, client)
	if err != nil {
		return nil, err
	}

	return kubernetes.New(newRestClientWrap(c, pluginConfig.ServiceAccountJWT)), nil

	// clientConfig := client.Config{}
	// info := &auth.Info{}
	// info.BearerToken = kubeConfig.jwt

	// clientConfig, _ = info.MergeWithConfig(clientConfig)

	// return kubernetes.New(client.New(clientConfig)), nil
}

//echo -n | openssl s_client -connect 127.0.0.1:57571 | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p'
