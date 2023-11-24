package apphostgae

import (
	"context"
	"github.com/strongo/strongoapp"
	"google.golang.org/appengine/v2"
	"net/http"
	"strings"
)

func NewHttpAppHostGAE() strongoapp.HttpAppHost {
	return httpAppHostGae{}
}

var _ strongoapp.HttpAppHost = (*httpAppHostGae)(nil)

type httpAppHostGae struct {
}

var gaeDefaultVersionHostname = appengine.DefaultVersionHostname

func (h httpAppHostGae) GetEnvironment(c context.Context, _ *http.Request) strongoapp.Environment {
	hostname := gaeDefaultVersionHostname(c)
	return getEnvFromHost(hostname)
}

var newContext = appengine.NewContext

func (h httpAppHostGae) HandleWithContext(handler strongoapp.HttpHandlerWithContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := newContext(r)
		handler(c, w, r)
	}
}

func getEnvFromHost(host string) strongoapp.Environment {
	if strings.Contains(host, "dev") && strings.HasSuffix(host, ".appspot.com") {
		return strongoapp.EnvDevTest
	} else if host == "localhost" || strings.HasPrefix(host, "localhost:") || strings.HasSuffix(host, ".ngrok.io") || strings.Contains(host, "local") {
		return strongoapp.EnvLocal
	}
	return strongoapp.EnvProduction
}
