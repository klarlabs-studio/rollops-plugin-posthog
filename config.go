package posthog

import "os"

// FromEnv builds a Provider from the plugin's environment. Secrets and endpoint
// come from the plugin process, never from the Rollops target spec (Rollops
// passes only the flag name, environment, and percentage).
//
//	POSTHOG_API_URL     base URL (default https://us.posthog.com)
//	POSTHOG_TOKEN       personal API key (required)
//	POSTHOG_PROJECT_ID  numeric project id (required)
func FromEnv() Provider {
	base := os.Getenv("POSTHOG_API_URL")
	if base == "" {
		base = "https://us.posthog.com"
	}
	return Provider{
		BaseURL:   base,
		Token:     os.Getenv("POSTHOG_TOKEN"),
		ProjectID: os.Getenv("POSTHOG_PROJECT_ID"),
	}
}
