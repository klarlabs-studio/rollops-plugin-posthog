// Command rollops-plugin-posthog is a Rollops feature-flag provider plugin
// backed by PostHog. Build it, pin its sha256, and point a rollout's
// featureFlags.plugin at the binary.
package main

import (
	"fmt"
	"os"

	posthog "github.com/klarlabs-studio/rollops-plugin-posthog"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

func main() {
	safety := plugin.Safety{
		NetworkHosts: []string{"us.posthog.com:443"},
		EnvVars:      []string{"POSTHOG_API_URL", "POSTHOG_TOKEN", "POSTHOG_PROJECT_ID"},
		RiskClass:    plugin.RiskActive,
	}
	if err := plugin.ServeFlagProvider("klarlabs/posthog", version, posthog.FromEnv(), safety); err != nil {
		fmt.Fprintln(os.Stderr, "rollops-plugin-posthog:", err)
		os.Exit(1)
	}
}
