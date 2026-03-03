package cmdutil

import (
	"context"
	"fmt"
	"testing"

	commonv1 "buf.build/gen/go/admiral/common/protocolbuffers/go/admiral/common/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

// ---------------------------------------------------------------------------
// Mock clients
// ---------------------------------------------------------------------------

type mockAppClient struct {
	applicationv1.ApplicationAPIClient // embed for unused methods
	resp                               *applicationv1.ListApplicationsResponse
	err                                error
}

func (m *mockAppClient) ListApplications(_ context.Context, _ *applicationv1.ListApplicationsRequest, _ ...grpc.CallOption) (*applicationv1.ListApplicationsResponse, error) {
	return m.resp, m.err
}

type mockEnvClient struct {
	environmentv1.EnvironmentAPIClient
	resp *environmentv1.ListEnvironmentsResponse
	err  error
}

func (m *mockEnvClient) ListEnvironments(_ context.Context, _ *environmentv1.ListEnvironmentsRequest, _ ...grpc.CallOption) (*environmentv1.ListEnvironmentsResponse, error) {
	return m.resp, m.err
}

type mockClusterClient struct {
	clusterv1.ClusterAPIClient
	resp *clusterv1.ListClustersResponse
	err  error
}

func (m *mockClusterClient) ListClusters(_ context.Context, _ *clusterv1.ListClustersRequest, _ ...grpc.CallOption) (*clusterv1.ListClustersResponse, error) {
	return m.resp, m.err
}

type mockClusterTokenClient struct {
	clusterv1.ClusterAPIClient
	resp *clusterv1.ListClusterTokensResponse
	err  error
}

func (m *mockClusterTokenClient) ListClusterTokens(_ context.Context, _ *clusterv1.ListClusterTokensRequest, _ ...grpc.CallOption) (*clusterv1.ListClusterTokensResponse, error) {
	return m.resp, m.err
}

// ---------------------------------------------------------------------------
// ResolveAppID
// ---------------------------------------------------------------------------

func TestResolveAppID(t *testing.T) {
	ctx := context.Background()

	t.Run("idFlag set returns directly", func(t *testing.T) {
		id, err := ResolveAppID(ctx, nil, "", "some-uuid")
		require.NoError(t, err)
		require.Equal(t, "some-uuid", id)
	})

	t.Run("name found returns ID", func(t *testing.T) {
		client := &mockAppClient{
			resp: &applicationv1.ListApplicationsResponse{
				Applications: []*applicationv1.Application{
					{Id: "app-123", Name: "billing-api"},
				},
			},
		}
		id, err := ResolveAppID(ctx, client, "billing-api", "")
		require.NoError(t, err)
		require.Equal(t, "app-123", id)
	})

	t.Run("name not found", func(t *testing.T) {
		client := &mockAppClient{
			resp: &applicationv1.ListApplicationsResponse{},
		}
		_, err := ResolveAppID(ctx, client, "ghost", "")
		require.ErrorContains(t, err, `application "ghost" not found`)
	})

	t.Run("multiple matches", func(t *testing.T) {
		client := &mockAppClient{
			resp: &applicationv1.ListApplicationsResponse{
				Applications: []*applicationv1.Application{
					{Id: "a1"}, {Id: "a2"},
				},
			},
		}
		_, err := ResolveAppID(ctx, client, "dup", "")
		require.ErrorContains(t, err, "multiple applications match")
	})

	t.Run("empty name and empty idFlag", func(t *testing.T) {
		_, err := ResolveAppID(ctx, nil, "", "")
		require.ErrorContains(t, err, "no application name")
	})

	t.Run("RPC error", func(t *testing.T) {
		client := &mockAppClient{err: fmt.Errorf("connection refused")}
		_, err := ResolveAppID(ctx, client, "billing-api", "")
		require.ErrorContains(t, err, "looking up application")
		require.ErrorContains(t, err, "connection refused")
	})
}

// ---------------------------------------------------------------------------
// ResolveEnvID
// ---------------------------------------------------------------------------

func TestResolveEnvID(t *testing.T) {
	ctx := context.Background()

	t.Run("idFlag set returns directly", func(t *testing.T) {
		id, err := ResolveEnvID(ctx, nil, "app-1", "", "env-uuid")
		require.NoError(t, err)
		require.Equal(t, "env-uuid", id)
	})

	t.Run("name found returns ID", func(t *testing.T) {
		client := &mockEnvClient{
			resp: &environmentv1.ListEnvironmentsResponse{
				Environments: []*environmentv1.Environment{
					{Id: "env-123", Name: "staging"},
				},
			},
		}
		id, err := ResolveEnvID(ctx, client, "app-1", "staging", "")
		require.NoError(t, err)
		require.Equal(t, "env-123", id)
	})

	t.Run("name not found", func(t *testing.T) {
		client := &mockEnvClient{
			resp: &environmentv1.ListEnvironmentsResponse{},
		}
		_, err := ResolveEnvID(ctx, client, "app-1", "ghost", "")
		require.ErrorContains(t, err, `environment "ghost" not found`)
	})

	t.Run("multiple matches", func(t *testing.T) {
		client := &mockEnvClient{
			resp: &environmentv1.ListEnvironmentsResponse{
				Environments: []*environmentv1.Environment{
					{Id: "e1"}, {Id: "e2"},
				},
			},
		}
		_, err := ResolveEnvID(ctx, client, "app-1", "dup", "")
		require.ErrorContains(t, err, "multiple environments match")
	})

	t.Run("empty name and empty idFlag", func(t *testing.T) {
		_, err := ResolveEnvID(ctx, nil, "app-1", "", "")
		require.ErrorContains(t, err, "no environment name")
	})

	t.Run("RPC error", func(t *testing.T) {
		client := &mockEnvClient{err: fmt.Errorf("timeout")}
		_, err := ResolveEnvID(ctx, client, "app-1", "staging", "")
		require.ErrorContains(t, err, "looking up environment")
		require.ErrorContains(t, err, "timeout")
	})
}

// ---------------------------------------------------------------------------
// ResolveClusterID
// ---------------------------------------------------------------------------

func TestResolveClusterID(t *testing.T) {
	ctx := context.Background()

	t.Run("idFlag set returns directly", func(t *testing.T) {
		id, err := ResolveClusterID(ctx, nil, "", "cl-uuid")
		require.NoError(t, err)
		require.Equal(t, "cl-uuid", id)
	})

	t.Run("name found returns ID", func(t *testing.T) {
		client := &mockClusterClient{
			resp: &clusterv1.ListClustersResponse{
				Clusters: []*clusterv1.Cluster{
					{Id: "cl-123", Name: "prod-cluster"},
				},
			},
		}
		id, err := ResolveClusterID(ctx, client, "prod-cluster", "")
		require.NoError(t, err)
		require.Equal(t, "cl-123", id)
	})

	t.Run("name not found", func(t *testing.T) {
		client := &mockClusterClient{
			resp: &clusterv1.ListClustersResponse{},
		}
		_, err := ResolveClusterID(ctx, client, "ghost", "")
		require.ErrorContains(t, err, `cluster "ghost" not found`)
	})

	t.Run("multiple matches", func(t *testing.T) {
		client := &mockClusterClient{
			resp: &clusterv1.ListClustersResponse{
				Clusters: []*clusterv1.Cluster{
					{Id: "c1"}, {Id: "c2"},
				},
			},
		}
		_, err := ResolveClusterID(ctx, client, "dup", "")
		require.ErrorContains(t, err, "multiple clusters match")
	})

	t.Run("empty name and empty idFlag", func(t *testing.T) {
		_, err := ResolveClusterID(ctx, nil, "", "")
		require.ErrorContains(t, err, "no cluster name")
	})

	t.Run("RPC error", func(t *testing.T) {
		client := &mockClusterClient{err: fmt.Errorf("unavailable")}
		_, err := ResolveClusterID(ctx, client, "prod", "")
		require.ErrorContains(t, err, "looking up cluster")
		require.ErrorContains(t, err, "unavailable")
	})
}

// ---------------------------------------------------------------------------
// ResolveClusterTokenID
// ---------------------------------------------------------------------------

func TestResolveClusterTokenID(t *testing.T) {
	ctx := context.Background()

	t.Run("idFlag set returns directly", func(t *testing.T) {
		id, err := ResolveClusterTokenID(ctx, nil, "cl-1", "", "tok-uuid")
		require.NoError(t, err)
		require.Equal(t, "tok-uuid", id)
	})

	t.Run("name found returns ID", func(t *testing.T) {
		client := &mockClusterTokenClient{
			resp: &clusterv1.ListClusterTokensResponse{
				AccessTokens: []*commonv1.AccessToken{
					{Id: "tok-123", Name: "ci-deploy-key"},
				},
			},
		}
		id, err := ResolveClusterTokenID(ctx, client, "cl-1", "ci-deploy-key", "")
		require.NoError(t, err)
		require.Equal(t, "tok-123", id)
	})

	t.Run("client-side filter ignores non-matching names", func(t *testing.T) {
		client := &mockClusterTokenClient{
			resp: &clusterv1.ListClusterTokensResponse{
				AccessTokens: []*commonv1.AccessToken{
					{Id: "tok-1", Name: "bar"},
					{Id: "tok-2", Name: "default"},
				},
			},
		}
		id, err := ResolveClusterTokenID(ctx, client, "cl-1", "bar", "")
		require.NoError(t, err)
		require.Equal(t, "tok-1", id)
	})

	t.Run("name not found", func(t *testing.T) {
		client := &mockClusterTokenClient{
			resp: &clusterv1.ListClusterTokensResponse{},
		}
		_, err := ResolveClusterTokenID(ctx, client, "cl-1", "ghost", "")
		require.ErrorContains(t, err, `token "ghost" not found`)
	})

	t.Run("multiple matches", func(t *testing.T) {
		client := &mockClusterTokenClient{
			resp: &clusterv1.ListClusterTokensResponse{
				AccessTokens: []*commonv1.AccessToken{
					{Id: "t1", Name: "dup"}, {Id: "t2", Name: "dup"},
				},
			},
		}
		_, err := ResolveClusterTokenID(ctx, client, "cl-1", "dup", "")
		require.ErrorContains(t, err, "multiple tokens match")
	})

	t.Run("empty name and empty idFlag", func(t *testing.T) {
		_, err := ResolveClusterTokenID(ctx, nil, "cl-1", "", "")
		require.ErrorContains(t, err, "no token name")
	})

	t.Run("RPC error", func(t *testing.T) {
		client := &mockClusterTokenClient{err: fmt.Errorf("permission denied")}
		_, err := ResolveClusterTokenID(ctx, client, "cl-1", "my-token", "")
		require.ErrorContains(t, err, "looking up token")
		require.ErrorContains(t, err, "permission denied")
	})
}
