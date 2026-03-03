package cmdutil

import (
	"context"
	"fmt"

	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

// ResolveAppID resolves an application identifier to its UUID.
// If idFlag is set, it is returned directly. Otherwise, the name is looked up
// via the list endpoint with a name filter.
func ResolveAppID(ctx context.Context, appClient applicationv1.ApplicationAPIClient, name, idFlag string) (string, error) {
	if idFlag != "" {
		return idFlag, nil
	}
	if name == "" {
		return "", fmt.Errorf("no application name or ID provided")
	}

	resp, err := appClient.ListApplications(ctx, &applicationv1.ListApplicationsRequest{
		Filter: fmt.Sprintf("field['name'] = '%s'", name),
	})
	if err != nil {
		return "", fmt.Errorf("looking up application %q: %w", name, err)
	}

	switch len(resp.Applications) {
	case 0:
		return "", fmt.Errorf("application %q not found", name)
	case 1:
		return resp.Applications[0].Id, nil
	default:
		return "", fmt.Errorf("multiple applications match name %q; use --id to specify", name)
	}
}

// ResolveEnvID resolves an environment name to its UUID within an application.
// If idFlag is set, it is returned directly. Otherwise, the name is looked up
// via the list endpoint filtered by application_id and name.
func ResolveEnvID(ctx context.Context, envClient environmentv1.EnvironmentAPIClient, appID, name, idFlag string) (string, error) {
	if idFlag != "" {
		return idFlag, nil
	}
	if name == "" {
		return "", fmt.Errorf("no environment name or ID provided")
	}

	resp, err := envClient.ListEnvironments(ctx, &environmentv1.ListEnvironmentsRequest{
		Filter: fmt.Sprintf("field['application_id'] = '%s' AND field['name'] = '%s'", appID, name),
	})
	if err != nil {
		return "", fmt.Errorf("looking up environment %q: %w", name, err)
	}

	switch len(resp.Environments) {
	case 0:
		return "", fmt.Errorf("environment %q not found", name)
	case 1:
		return resp.Environments[0].Id, nil
	default:
		return "", fmt.Errorf("multiple environments match name %q; use --id to specify", name)
	}
}

// ResolveClusterTokenID resolves a cluster token identifier to its UUID.
// If idFlag is set, it is returned directly. Otherwise, the name is looked up
// via the list endpoint with a name filter (with client-side fallback).
func ResolveClusterTokenID(ctx context.Context, clusterClient clusterv1.ClusterAPIClient, clusterID, name, idFlag string) (string, error) {
	if idFlag != "" {
		return idFlag, nil
	}
	if name == "" {
		return "", fmt.Errorf("no token name or ID provided")
	}

	resp, err := clusterClient.ListClusterTokens(ctx, &clusterv1.ListClusterTokensRequest{
		ClusterId: clusterID,
		Filter:    fmt.Sprintf("field['name'] = '%s'", name),
	})
	if err != nil {
		return "", fmt.Errorf("looking up token %q: %w", name, err)
	}

	// Client-side filter: the API may not apply the name filter.
	var matched []string
	for _, t := range resp.AccessTokens {
		if t.Name == name {
			matched = append(matched, t.Id)
		}
	}

	switch len(matched) {
	case 0:
		return "", fmt.Errorf("token %q not found", name)
	case 1:
		return matched[0], nil
	default:
		return "", fmt.Errorf("multiple tokens match name %q; use --token-id to specify", name)
	}
}

// ResolveClusterID resolves a cluster identifier to its UUID.
// If idFlag is set, it is returned directly. Otherwise, the name is looked up
// via the list endpoint with a name filter.
func ResolveClusterID(ctx context.Context, clusterClient clusterv1.ClusterAPIClient, name, idFlag string) (string, error) {
	if idFlag != "" {
		return idFlag, nil
	}
	if name == "" {
		return "", fmt.Errorf("no cluster name or ID provided")
	}

	resp, err := clusterClient.ListClusters(ctx, &clusterv1.ListClustersRequest{
		Filter: fmt.Sprintf("field['name'] = '%s'", name),
	})
	if err != nil {
		return "", fmt.Errorf("looking up cluster %q: %w", name, err)
	}

	switch len(resp.Clusters) {
	case 0:
		return "", fmt.Errorf("cluster %q not found", name)
	case 1:
		return resp.Clusters[0].Id, nil
	default:
		return "", fmt.Errorf("multiple clusters match name %q; use --id to specify", name)
	}
}
