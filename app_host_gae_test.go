package apphostgae

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/strongo/strongoapp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHttpAppHostGAE(t *testing.T) {

	httpAppHost := func() strongoapp.HttpAppHost {
		return NewHttpAppHostGAE()
	}()
	assert.NotNil(t, httpAppHost)
}

func Test_getEnvFromHost(t *testing.T) {
	// Covered by Test_httpAppHostGae_GetEnvironment
}

func Test_httpAppHostGae_GetEnvironment(t *testing.T) {
	type args struct {
		c context.Context
		r *http.Request
	}
	tests := []struct {
		name string
		host string
		args args
		want strongoapp.Environment
	}{
		{"appspot", "some-app.appspot.com", args{context.Background(), nil}, strongoapp.EnvProduction},
		{"local", "some-app.local", args{context.Background(), nil}, strongoapp.EnvLocal},
		{"localhost", "localhost", args{context.Background(), nil}, strongoapp.EnvLocal},
	}
	h := httpAppHostGae{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gaeDefaultVersionHostname = func(c context.Context) string {
				return tt.host
			}
			environment := h.GetEnvironment(context.Background(), nil)
			assert.Equal(t, tt.want, environment)
		})
	}
}

func Test_httpAppHostGae_HandleWithContext(t *testing.T) {
	type args struct {
		handler strongoapp.HttpHandlerWithContext
	}
	tests := []struct {
		name   string
		status int
		args   args
	}{
		{"should_pass", http.StatusOK, args{func(c context.Context, w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}}},
	}
	h := httpAppHostGae{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := h.HandleWithContext(tt.args.handler)
			assert.NotNil(t, handler)
			newContextCalled := false
			request := httptest.NewRequest("GET", "/", nil)
			newContext = func(r *http.Request) context.Context {
				assert.Equal(t, r, request)
				newContextCalled = true
				return context.Background()
			}
			responseRecorder := httptest.NewRecorder()
			handler(responseRecorder, request)
			assert.True(t, newContextCalled)
			assert.Equal(t, tt.status, responseRecorder.Code)
		})
	}
}
