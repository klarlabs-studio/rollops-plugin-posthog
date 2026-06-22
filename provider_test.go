package posthog

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.klarlabs.de/rollops/pkg/plugin"
)

func TestApplyFlag_ResolvesAndPatches(t *testing.T) {
	var patchBody map[string]any
	var patchedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(flagList{Results: []flag{{ID: 7, Key: "checkout"}}})
		case http.MethodPatch:
			patchedPath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &patchBody)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "k", ProjectID: "42", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "checkout", Environment: "prod", Percentage: 30}); err != nil {
		t.Fatalf("ApplyFlag: %v", err)
	}
	if patchedPath != "/api/projects/42/feature_flags/7/" {
		t.Errorf("patched wrong path: %s", patchedPath)
	}
	if patchBody["active"] != true {
		t.Errorf("active = %v, want true", patchBody["active"])
	}
	filters, _ := patchBody["filters"].(map[string]any)
	groups, _ := filters["groups"].([]any)
	g0, _ := groups[0].(map[string]any)
	if g0["rollout_percentage"].(float64) != 30 {
		t.Errorf("rollout_percentage = %v, want 30", g0["rollout_percentage"])
	}
}

func TestApplyFlag_Paginates(t *testing.T) {
	var page int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			w.WriteHeader(200)
			return
		}
		page++
		if page == 1 {
			// First page lacks the flag and points to a second page.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []flag{{ID: 1, Key: "other"}},
				"next":    "http://" + r.Host + "/api/projects/42/feature_flags/?page=2",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(flagList{Results: []flag{{ID: 9, Key: "checkout"}}})
	}))
	defer srv.Close()

	p := Provider{BaseURL: srv.URL, Token: "k", ProjectID: "42", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "checkout", Environment: "prod", Percentage: 10}); err != nil {
		t.Fatalf("ApplyFlag across pages: %v", err)
	}
	if page < 2 {
		t.Errorf("expected pagination to fetch a second page, pages=%d", page)
	}
}

func TestApplyFlag_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(flagList{Results: []flag{{ID: 1, Key: "other"}}})
	}))
	defer srv.Close()
	p := Provider{BaseURL: srv.URL, Token: "k", ProjectID: "42", HTTP: srv.Client()}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "missing", Environment: "p"}); err == nil {
		t.Fatal("unknown flag must error")
	}
}

func TestApplyFlag_RequiresCreds(t *testing.T) {
	p := Provider{BaseURL: "http://x"}
	if err := p.ApplyFlag(context.Background(), plugin.FlagChange{Flag: "f"}); err == nil {
		t.Fatal("missing creds must error")
	}
}
