# Design: Ephemeral Environments & Promotion Order

**Status**: Proposal — not yet in proto
**Origin**: Concepts discovered during CLI stub implementation, deferred when aligning with actual Environment proto

---

## Problem

Teams need two capabilities that the current Environment proto doesn't support:

1. **Short-lived preview/PR environments** that auto-expire after a TTL
2. **Explicit promotion ordering** between environments (dev → staging → prod)

Both are core to Admiral's design principles ("promotion is explicit") and competitive positioning (vs. Humanitec's Score-based approach).

---

## Ephemeral Environments

### Concept

An ephemeral environment is a time-bounded deployment target, typically created per pull request or feature branch, that is automatically cleaned up after a configurable TTL expires.

### Proposed Proto Changes

```protobuf
enum LifecycleType {
  LIFECYCLE_TYPE_UNSPECIFIED = 0;
  LIFECYCLE_TYPE_PERMANENT  = 1;  // Default. Lives until explicitly deleted.
  LIFECYCLE_TYPE_EPHEMERAL  = 2;  // Auto-expires after TTL.
}

message Environment {
  // ... existing fields ...

  // Whether this environment is permanent or ephemeral.
  LifecycleType lifecycle_type = 16;

  // Time-to-live for ephemeral environments. Ignored for permanent environments.
  // Server starts the countdown from creation time. When TTL expires, the server
  // triggers a destroy deployment and then deletes the environment record.
  google.protobuf.Duration ttl = 17;

  // The permanent environment this ephemeral env is linked to.
  // Used for inheriting config defaults and tracking lineage.
  // Must reference an environment in the same application with
  // lifecycle_type = PERMANENT.
  string parent_environment_id = 18;

  // Source reference (branch, tag, or commit SHA) this environment tracks.
  // Informational — Admiral does not auto-deploy on ref changes (that's the
  // CI/CD system's job), but stores this for display, filtering, and
  // traceability.
  string source_ref = 19;
}
```

### CLI Flags

```
admiral env create preview-123 \
  --app billing-api \
  --lifecycle ephemeral \
  --ttl 24h \
  --parent staging \
  --source-ref feature/payments-v2 \
  --cluster dev-cluster \
  --runtime-type kubernetes
```

### Behavior

| Aspect | Permanent | Ephemeral |
|--------|-----------|-----------|
| Deletion | Manual (`--confirm`) | Auto after TTL or manual |
| Config inheritance | None | From parent env |
| Typical count per app | 2-5 | 0-N (scales with PRs) |
| TTL | N/A | Required |
| Parent | N/A | Optional (recommended) |

### Server Responsibilities

- Validate: `ttl` must be set when `lifecycle_type = EPHEMERAL`, rejected otherwise
- Validate: `parent_environment_id` must reference a `PERMANENT` env in the same app
- Background job: check expired ephemeral environments, trigger destroy + delete
- Expose `lifecycle_type` and `ttl_remaining` in list/get responses for visibility

### Open Questions

1. **TTL extension**: Should users be able to extend TTL on a running ephemeral env? (`admiral env update preview-123 --ttl 48h`)
2. **Destroy semantics**: When TTL expires, should Admiral destroy all components first (graceful) or just delete the environment record (hard)? Graceful is safer but slower.
3. **Notifications**: Should Admiral notify (webhook/event) before TTL expiry? (e.g., 1h warning)
4. **Config inheritance**: How deep does parent inheritance go? Just variables? Also component overrides?

---

## Promotion Order

### Concept

Each environment has an integer `promotion_order` that defines its position in the promotion chain. Lower numbers are promoted first. This enables Admiral to:

- Enforce promotion policies (can't deploy to prod before staging)
- Suggest next promotion target in UI/CLI
- Visualize the promotion pipeline

### Proposed Proto Changes

```protobuf
message Environment {
  // ... existing fields ...

  // Position in the promotion chain. Lower values are promoted first.
  // Convention: dev=10, staging=50, prod=100. Gaps allow inserting
  // environments later without renumbering.
  // Zero means "no explicit order" (environment is outside the promotion chain).
  int32 promotion_order = 20;
}
```

### CLI Flags

```
admiral env create staging \
  --app billing-api \
  --promotion-order 50 \
  --runtime-type kubernetes \
  --cluster staging-cluster
```

### Behavior

- `promotion_order = 0` means the environment is not part of the ordered chain (e.g., ephemeral envs, sandbox envs)
- The server does **not** enforce promotion order by default (teams opt in via policy)
- `admiral env list` could show environments sorted by promotion order
- Future: `admiral env promote staging --to production` could use promotion order to validate the transition

### Promotion Policy (Future)

This is out of scope for the initial implementation but worth designing for:

```
admiral app set-policy billing-api --require-promotion-order
```

When enabled, deployments to an environment are rejected unless all lower-order environments have a successful deployment of the same (or newer) artifact version.

### Open Questions

1. **Enforcement**: Should promotion order be advisory-only initially, or should we ship with optional enforcement?
2. **Naming**: `promotion_order` vs `deploy_order` vs `stage_order`?
3. **Visualization**: Should `env list` default-sort by promotion order when set?

---

## Implementation Priority

| Feature | Effort | Value | Suggested Priority |
|---------|--------|-------|--------------------|
| `promotion_order` field | Small | Medium | P1 — simple to add, immediately useful |
| `lifecycle_type` + `ttl` | Medium | High | P1 — differentiator vs competitors |
| `parent_environment_id` | Small | Medium | P2 — useful but not blocking |
| `source_ref` | Small | Low-Medium | P2 — nice for traceability |
| TTL expiry background job | Medium | High | P1 — required for ephemeral to work |
| Promotion policy enforcement | Large | Medium | P3 — can ship advisory-only first |

| Runtime model normalization | Medium | High | P2 — structural improvement, enables multi-runtime |

---

## Runtime Model Normalization

**Status**: Future consideration

The current Environment proto is asymmetric — workload has type + config, infra has config only. As Admiral supports more runtimes, both should follow a consistent `type + config oneof` pattern.

### Known Runtime Types

| Category | Type | Config Fields |
|----------|------|---------------|
| **Workload** | Kubernetes | cluster_id, namespace |
| **Workload** | Cloud Run | project, region, service |
| **Workload** | ECS | cluster, service, task definition |
| **Infrastructure** | Terraform / OpenTofu | runner_id |
| **Infrastructure** | Pulumi | stack, project, backend |
| **Infrastructure** | CloudFormation | stack name, region |

### Proposed Proto Shape

```protobuf
enum WorkloadType {
  WORKLOAD_TYPE_UNSPECIFIED = 0;
  WORKLOAD_TYPE_KUBERNETES  = 1;
  WORKLOAD_TYPE_CLOUD_RUN   = 2;
  WORKLOAD_TYPE_ECS         = 3;
}

enum InfraType {
  INFRA_TYPE_UNSPECIFIED     = 0;
  INFRA_TYPE_TERRAFORM       = 1;  // Covers OpenTofu
  INFRA_TYPE_PULUMI          = 2;
  INFRA_TYPE_CLOUDFORMATION  = 3;
}

message Environment {
  WorkloadType workload_type = ...;
  oneof workload_config {
    KubernetesConfig kubernetes = ...;
    CloudRunConfig   cloud_run  = ...;
    ECSConfig        ecs        = ...;
  }

  InfraType infra_type = ...;
  oneof infra_config {
    TerraformConfig      terraform      = ...;
    PulumiConfig         pulumi         = ...;
    CloudFormationConfig cloudformation = ...;
  }
}
```

### CLI Impact

Current flags (`--runtime-type`, `--cluster`, `--namespace`, `--runner`) would evolve to:
```
--workload-type kubernetes --cluster prod --namespace app
--infra-type terraform --runner runner-01
```

This is a breaking change to the proto and CLI flags, so it should be done before GA.

---

## References

- [Admiral Design Principles](../../CLAUDE.md) — "promotion is explicit"
- [API/UX Feedback Notes](../../.claude/projects/-Users-mberwanger-Development-datalift-admiral-cli/memory/api-ux-feedback.md) — original discovery context
- Humanitec Score spec — competitor reference for environment lifecycle
