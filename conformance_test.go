package posthog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.klarlabs.de/rollops/pkg/flagconformance"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// fakePostHog returns a flag list containing the sample key, then accepts PATCH.
func fakePostHog(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte(`{"results":[{"id":7,"key":"checkout"}]}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestConformance(t *testing.T) {
	flagconformance.Run(t, func() (plugin.FlagProvider, error) {
		srv := fakePostHog(t)
		return Provider{BaseURL: srv.URL, Token: "k", ProjectID: "42", HTTP: srv.Client()}, nil
	}, plugin.FlagChange{Flag: "checkout", Environment: "production"})
}
