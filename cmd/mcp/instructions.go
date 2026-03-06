package mcp

const admiralInstructions = `Admiral is a platform orchestrator that manages application deployments across infrastructure.

## Domain Model

Organization (implicit, derived from auth context)
├── Applications — logical service boundaries (e.g. "billing-api", "auth-service")
│   └── Environments — deployment targets within an app (e.g. "staging", "production")
├── Clusters — registered Kubernetes clusters that environments deploy to
│   ├── Cluster Tokens — authentication tokens for cluster agents
│   └── Workloads — running workloads discovered on the cluster
├── Variables — configuration key-value pairs with three scope levels:
│   ├── Global — no app or env specified; available to all
│   ├── App-scoped — app specified, no env; inherited by all envs in that app
│   └── Env-scoped — app + env specified; most specific, overrides inherited values
└── Personal Access Tokens — API tokens for the current user

All names are unique within their parent scope. Names can be used interchangeably
with UUIDs — the server resolves names to IDs automatically.

## Tool Overview

- admiral_query: Read-only access to all resources. Uses "resource" + "action" (list/get). Action defaults to "list" if omitted.
- admiral_set_variable: Create or update a variable (upsert). Scope is set by which of app/env you provide.
- admiral_delete_variable: Permanently delete a variable by key+scope or by UUID.

## admiral_query Required Fields by Resource

| resource.action       | required fields             | notes                              |
|-----------------------|-----------------------------|------------------------------------|
| app.list              | (none)                      |                                    |
| app.get               | app or id                   |                                    |
| env.list              | app                         | lists envs within that app         |
| env.get               | app + env, or id            |                                    |
| variable.list         | (none)                      | omit app/env for global only       |
| variable.get          | key + scope, or id          | scope = app and/or env             |
| cluster.list          | (none)                      |                                    |
| cluster.get           | cluster or cluster_id or id |                                    |
| cluster_status.get    | cluster or cluster_id or id | health/connectivity info           |
| cluster_token.list    | cluster or cluster_id       |                                    |
| cluster_token.get     | cluster or cluster_id + id  | id = token UUID                    |
| workload.list         | cluster or cluster_id       | list only, no get                  |
| token.list            | (none)                      | personal access tokens             |
| token.get             | id                          | token UUID required                |
| whoami.get            | (none)                      | current user, org, and auth info   |

Not all resources support both actions. If an action is not listed above,
it is not supported (e.g. workload.get, cluster_status.list do not exist).

## Pagination

List results default to 50 items per page. Use page_size to adjust (max varies by resource).
If next_page_token is present in the response, pass it back as page_token to fetch the next page.
Repeat until next_page_token is empty.

## Filtering Lists

List actions accept an optional "filter" field using Admiral's filter DSL:

  field['fieldname'] = 'value'

Operators: =, !=, <, >, <=, >=, ~= (regex match)
Logical:   AND, OR, NOT
Predicates: IN, BETWEEN, CONTAINS, STARTS_WITH, ENDS_WITH, IS NULL, EXISTS

Filterable fields by resource:
- app: name, labels.<key>
- env: name, runtime_type, labels.<key> (application_id is set automatically from "app")
- variable: key, type, sensitive (scope is set automatically from app/env fields)
- cluster: name, health_status, labels.<key>
- workload: namespace, kind, name, health_status
- cluster_token: name, status (ACTIVE, REVOKED)
- token: name, status

Examples:
  field['name'] = 'billing-api'
  field['labels.region'] = 'us-east-1' AND field['labels.env'] = 'prod'
  field['health_status'] = 'DEGRADED'
  field['namespace'] = 'production' AND field['kind'] = 'Deployment'

Note: for env and variable lists, scope filters (application_id, environment_id)
are applied automatically based on the app/env input fields. The filter field
is for additional filtering on top of that.

## Variable Scoping

When listing variables, the scope filters control what is returned:
- No app/env: returns global variables only
- app only: returns variables scoped to that application
- app + env: returns variables scoped to that specific environment

Variable inheritance (global -> app -> env) is resolved server-side at deployment time,
not in list responses.

## admiral_set_variable Fields

- key (required): variable name (e.g. IMAGE_TAG, DB_URL)
- value (required): variable value
- app: application name (omit for global scope)
- env: environment name (requires app)
- sensitive: encrypt at rest and mask in responses
- type: value type — string (default), number, boolean, or complex
- description: purpose of this variable (cannot be cleared once set)

## Response Format

All responses use proto3 JSON serialization with camelCase field names.
List responses contain an array field named after the resource (e.g. "applications",
"environments", "variables") plus a "nextPageToken" if more pages exist.
Get responses return the resource object directly.

## Common Workflows

1. Check auth: whoami.get to verify credentials and see current org
2. Explore: list apps -> list envs for an app -> list variables at each scope
3. Configure: set variables at the appropriate scope level
4. Inspect infrastructure: list clusters -> check cluster status -> list workloads
5. Audit: list personal access tokens, review cluster tokens

## Authentication

Uses the same credentials as the Admiral CLI. If calls fail with auth errors,
the user should run "admiral auth login" or check ADMIRAL_TOKEN.`
