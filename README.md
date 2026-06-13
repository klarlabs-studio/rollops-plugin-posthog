# rollops-plugin-posthog

A [Rollops](https://github.com/klarlabs-studio/rollops) feature-flag provider
plugin backed by [PostHog](https://posthog.com/). It drives a PostHog feature
flag's active state and rollout percentage to track a Rollops canary in
lockstep — as a rollout steps 10% → 50% → 100%, the flag follows.

## How it works

Rollops calls the plugin per progressive step (and/or on promote) with the flag
key, the target environment, and the current traffic percentage. The plugin
resolves the flag id by key (following pagination), then PATCHes its `active`
state and a single rollout group at the percentage.

In PostHog a project is the unit of environment isolation, so the plugin targets
the project set by `POSTHOG_PROJECT_ID`; the Rollops `environment` is
informational.

## Configuration

Credentials come from the plugin's own environment, never from the Rollops
target spec:

| Env var              | Required | Default                  | Description                  |
|----------------------|----------|--------------------------|------------------------------|
| `POSTHOG_API_URL`    | no       | `https://us.posthog.com` | Base URL (use the EU host if applicable) |
| `POSTHOG_TOKEN`      | yes      | —                        | Personal API key (`Bearer`)  |
| `POSTHOG_PROJECT_ID` | yes      | —                        | Numeric project id           |

## Install

```sh
rollops plugin install posthog
```

Or build and pin manually with `make build` / `make checksum`, then wire into a
rollout spec:

```yaml
featureFlags:
  plugin: ~/.rollops/plugins/posthog
  sha256: <pin>
  flag: checkout
  environment: production
  when: both
```

## License

MIT
