package apphostgae

import (
	"context"
	strongo "github.com/strongo/app"
	"google.golang.org/appengine/v2"
	"net/http"
	"strings"
)

func NewHttpAppHostGAE() strongo.HttpAppHost {
	return httpAppHostGae{}
}

var _ strongo.HttpAppHost = (*httpAppHostGae)(nil)

type httpAppHostGae struct {
}

var gaeDefaultVersionHostname = appengine.DefaultVersionHostname

func (h httpAppHostGae) GetEnvironment(c context.Context, _ *http.Request) strongo.Environment {
	hostname := gaeDefaultVersionHostname(c)
	return getEnvFromHost(hostname)
}

var newContext = appengine.NewContext

func (h httpAppHostGae) HandleWithContext(handler strongo.HttpHandlerWithContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := newContext(r)
		handler(c, w, r)
	}
}

func getEnvFromHost(host string) strongo.Environment {
	if strings.Contains(host, "dev") && strings.HasSuffix(host, ".appspot.com") {
		return strongo.EnvDevTest
	} else if host == "localhost" || strings.HasPrefix(host, "localhost:") || strings.HasSuffix(host, ".ngrok.io") || strings.Contains(host, "local") {
		return strongo.EnvLocal
	}
	return strongo.EnvProduction
}
