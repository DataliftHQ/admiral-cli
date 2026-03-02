# Design: App Status, Diff & Clone Commands

**Status**: Deferred — not in v1, pending API design
**Origin**: CLI stubs removed from command tree to avoid shipping unimplemented features

---

## `app status` — Deployment Status Across Environments

Show the deployment status of an application across all (or a filtered) environments.

### Usage

```
admiral app status [app] [-e/--env <env>]
```

### Examples

```bash
# Show status across all environments
admiral app status billing-api

# Show status for a specific environment
admiral app status billing-api -e staging

# Use the active app context
admiral use billing-api
admiral app status -e production
```

### Flags

| Flag | Type | Description |
|------|------|-------------|
| `-e`, `--env` | string | Filter to a specific environment |

### Expected Output

Tabular view showing each environment's deployment state:

```
ENVIRONMENT   STATUS      VERSION   DEPLOYED         AGE
dev           Healthy     v1.4.2    2025-01-15 10:30 5d
staging       Deploying   v1.4.3    2025-01-20 14:22 2m
production    Healthy     v1.4.1    2025-01-10 09:00 10d
```

### API Dependencies

- Needs a deployment status endpoint per environment (or aggregated across envs for an app)
- May depend on Deployment API (`deploymentv1`) being fully wired
- `has_pending_changes` on Environment is a partial signal but not deployment status

---

## `app diff` — Compare Configuration Between Environments

Compare an application's resolved configuration between two environments to see what differs before promoting.

### Usage

```
admiral app diff [app] --from <env> --to <env>
```

### Examples

```bash
# Compare staging and production
admiral app diff billing-api --from staging --to production

# Use the active app context
admiral use billing-api
admiral app diff --from staging --to production
```

### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--from` | string | Yes | Source environment |
| `--to` | string | Yes | Target environment |

### Expected Output

Diff-style output showing configuration differences:

```
Variables:
  + DATABASE_URL = postgres://prod-db:5432/app   (only in production)
  - DATABASE_URL = postgres://staging-db:5432/app (only in staging)
  ~ LOG_LEVEL: debug → info

Components:
  ~ api: image tag v1.4.2 → v1.4.1
  = worker: identical
```

### API Dependencies

- Needs resolved configuration per environment (variables, component settings, overrides)
- May need a dedicated diff endpoint or client-side comparison of two GetEnvironment responses
- Variable API (`variablev1`) must support listing resolved values per environment

---

## `app clone` — Clone Configuration Between Environments

Clone an application's configuration from one environment to another. Useful for setting up new environments or syncing config.

### Usage

```
admiral app clone [app] --from <env> --to <env> [flags]
```

### Examples

```bash
# Clone staging to production
admiral app clone billing-api --from staging --to production

# Clone including variables
admiral app clone billing-api --from staging --to production --include-variables

# Clone but exclude specific variables
admiral app clone billing-api --from staging --to production --include-variables --exclude-variable SECRET_KEY
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | (required) | Source environment |
| `--to` | string | (required) | Target environment |
| `--include-variables` | bool | false | Include variables in the clone |
| `--exclude-variable` | string[] | nil | Variable keys to exclude (repeatable) |

### Behavior

- By default clones component configuration only (not variables — they often contain environment-specific secrets)
- `--include-variables` opts in to variable cloning
- `--exclude-variable` allows cherry-picking which variables to skip (e.g., secrets, endpoints)
- Target environment must already exist
- Does NOT trigger a deployment — just copies configuration. User deploys separately.

### API Dependencies

- Needs ability to read resolved config from source environment
- Needs ability to write/overwrite config on target environment
- May be a dedicated server-side RPC (safer, atomic) or client-side read+write sequence
- Variable API must support bulk create/update per environment

---

## Implementation Priority

| Command | Complexity | Dependencies | Suggested Priority |
|---------|-----------|--------------|-------------------|
| `app status` | Medium | Deployment API | P1 — most commonly needed |
| `app diff` | Medium | Variable + Component resolution | P2 — key for promotion workflows |
| `app clone` | High | Variable API bulk ops, component config | P3 — convenience, can use diff + manual copy initially |

---

## Notes

- All three commands follow the standard app resolution pattern (positional arg or `admiral use` context)
- All three are cross-environment operations — they need the app resolved plus two environment references
- Consider whether these should live under `app` or under `env` (e.g., `env diff --from staging --to prod`)
- `status` could also become a top-level command: `admiral status` showing everything
