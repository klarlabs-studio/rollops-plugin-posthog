// Package posthog is a Rollops feature-flag provider plugin backed by PostHog's
// feature-flags API. It drives a flag's active state and its rollout percentage
// to match a rollout's progressive steps, so a PostHog flag tracks a Rollops
// canary in lockstep.
package posthog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.klarlabs.de/rollops/pkg/plugin"
)

// Provider talks to PostHog's API. BaseURL, Token, and ProjectID come from the
// plugin's environment (see Config). In PostHog a project is the unit of
// environment isolation; the Rollops environment is informational.
type Provider struct {
	BaseURL   string // e.g. https://us.posthog.com
	Token     string // personal API key (Authorization: Bearer <key>)
	ProjectID string // numeric PostHog project id
	HTTP      *http.Client
}

func (p Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

// ApplyFlag resolves the flag id by key, then PATCHes its active state and a
// single rollout group at the given percentage.
func (p Provider) ApplyFlag(ctx context.Context, c plugin.FlagChange) error {
	if p.Token == "" || p.ProjectID == "" {
		return fmt.Errorf("posthog: POSTHOG_TOKEN and POSTHOG_PROJECT_ID are required")
	}
	id, err := p.flagID(ctx, c.Flag)
	if err != nil {
		return err
	}
	return p.patch(ctx, id, !c.Disabled, c.Percentage)
}

type flag struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

type flagList struct {
	Results []flag `json:"results"`
	Next    string `json:"next"`
}

func (p Provider) flagID(ctx context.Context, key string) (int, error) {
	u := fmt.Sprintf("%s/api/projects/%s/feature_flags/", p.BaseURL, p.ProjectID)
	for u != "" {
		var out flagList
		if err := p.do(ctx, http.MethodGet, u, nil, &out); err != nil {
			return 0, fmt.Errorf("posthog: list feature flags: %w", err)
		}
		for _, f := range out.Results {
			if f.Key == key {
				return f.ID, nil
			}
		}
		u = out.Next // follow pagination
	}
	return 0, fmt.Errorf("posthog: flag %q not found in project %s", key, p.ProjectID)
}

func (p Provider) patch(ctx context.Context, id int, active bool, pct int) error {
	u := fmt.Sprintf("%s/api/projects/%s/feature_flags/%d/", p.BaseURL, p.ProjectID, id)
	body := map[string]any{
		"active": active,
		"filters": map[string]any{
			"groups": []map[string]any{{"properties": []any{}, "rollout_percentage": pct}},
		},
	}
	if err := p.do(ctx, http.MethodPatch, u, body, nil); err != nil {
		return fmt.Errorf("posthog: update flag %d: %w", id, err)
	}
	return nil
}

func (p Provider) do(ctx context.Context, method, u string, body, out any) error {
	var rdr *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
