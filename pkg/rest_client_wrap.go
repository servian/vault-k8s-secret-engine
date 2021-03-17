package servian

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

func newRestClientWrap(c *rest.RESTClient, jwt string) rest.Interface {
	wrap := restClientWrap{}
	wrap.restClient = c
	wrap.jwt = jwt

	return wrap
}

// restClientWrap is a thin wrapper around a rest.RESTClient that injects an authorization header to
//allow for Bearer token authentication using a JTW token for interation with the Kubernets API
type restClientWrap struct {
	restClient *rest.RESTClient
	jwt        string
}

func (c restClientWrap) APIVersion() schema.GroupVersion {
	return c.restClient.APIVersion()
}

func (c restClientWrap) GetRateLimiter() flowcontrol.RateLimiter {
	return c.restClient.GetRateLimiter()
}

func (c restClientWrap) Verb(verb string) *rest.Request {
	return c.insertAuthHeader(c.restClient.Verb(verb))
}

func (c restClientWrap) Post() *rest.Request {
	return c.insertAuthHeader(c.restClient.Post())
}

func (c restClientWrap) Put() *rest.Request {
	return c.insertAuthHeader(c.restClient.Put())
}

func (c restClientWrap) Patch(p types.PatchType) *rest.Request {
	return c.insertAuthHeader(c.restClient.Patch(p))
}

func (c restClientWrap) Get() *rest.Request {
	return c.insertAuthHeader(c.restClient.Get())
}

func (c restClientWrap) Delete() *rest.Request {
	return c.insertAuthHeader(c.restClient.Delete())
}

func (c restClientWrap) insertAuthHeader(r *rest.Request) *rest.Request {
	return r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
}
