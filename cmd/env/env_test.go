package env

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"go.admiral.io/cli/internal/factory"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

// ---------------------------------------------------------------------------
// resolveAppForEnv
// ---------------------------------------------------------------------------

func TestResolveAppForEnv(t *testing.T) {
	t.Run("flag set returns value", func(t *testing.T) {
		app, err := resolveAppForEnv("billing-api")
		require.NoError(t, err)
		require.Equal(t, "billing-api", app)
	})

	t.Run("empty flag errors", func(t *testing.T) {
		_, err := resolveAppForEnv("")
		require.ErrorContains(t, err, "no app specified")
	})
}

// ---------------------------------------------------------------------------
// parseRuntimeType
// ---------------------------------------------------------------------------

func TestParseRuntimeType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    environmentv1.RuntimeType
		wantErr string
	}{
		{
			name:  "kubernetes",
			input: "kubernetes",
			want:  environmentv1.RuntimeType_RUNTIME_TYPE_KUBERNETES,
		},
		{
			name:  "k8s alias",
			input: "k8s",
			want:  environmentv1.RuntimeType_RUNTIME_TYPE_KUBERNETES,
		},
		{
			name:  "case insensitive uppercase",
			input: "KUBERNETES",
			want:  environmentv1.RuntimeType_RUNTIME_TYPE_KUBERNETES,
		},
		{
			name:  "case insensitive mixed",
			input: "Kubernetes",
			want:  environmentv1.RuntimeType_RUNTIME_TYPE_KUBERNETES,
		},
		{
			name:    "invalid type",
			input:   "lambda",
			wantErr: `unsupported runtime type "lambda"`,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: `unsupported runtime type ""`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRuntimeType(tc.input)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// formatRuntimeConfig
// ---------------------------------------------------------------------------

func TestFormatRuntimeConfig(t *testing.T) {
	t.Run("nil env runtime config", func(t *testing.T) {
		env := &environmentv1.Environment{}
		require.Equal(t, "", formatRuntimeConfig(env))
	})

	t.Run("kubernetes cluster only", func(t *testing.T) {
		env := &environmentv1.Environment{
			RuntimeConfig: &environmentv1.Environment_Kubernetes{
				Kubernetes: &environmentv1.KubernetesConfig{
					ClusterId: "cl-abc",
				},
			},
		}
		require.Equal(t, "cluster=cl-abc", formatRuntimeConfig(env))
	})

	t.Run("kubernetes cluster and namespace", func(t *testing.T) {
		ns := "staging-ns"
		env := &environmentv1.Environment{
			RuntimeConfig: &environmentv1.Environment_Kubernetes{
				Kubernetes: &environmentv1.KubernetesConfig{
					ClusterId: "cl-abc",
					Namespace: &ns,
				},
			},
		}
		require.Equal(t, "cluster=cl-abc, namespace=staging-ns", formatRuntimeConfig(env))
	})

	t.Run("kubernetes with neither", func(t *testing.T) {
		env := &environmentv1.Environment{
			RuntimeConfig: &environmentv1.Environment_Kubernetes{
				Kubernetes: &environmentv1.KubernetesConfig{},
			},
		}
		require.Equal(t, "", formatRuntimeConfig(env))
	})

	t.Run("kubernetes namespace only", func(t *testing.T) {
		ns := "my-ns"
		env := &environmentv1.Environment{
			RuntimeConfig: &environmentv1.Environment_Kubernetes{
				Kubernetes: &environmentv1.KubernetesConfig{
					Namespace: &ns,
				},
			},
		}
		require.Equal(t, "namespace=my-ns", formatRuntimeConfig(env))
	})
}

// ---------------------------------------------------------------------------
// Command validation tests (no gRPC — errors fire before client creation)
// ---------------------------------------------------------------------------

func newTestEnvCmd(t *testing.T) *EnvCmd {
	t.Helper()
	return NewEnvCmd(&factory.Options{})
}

func TestCreateCmd_RequiresOneArg(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"create"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "requires 1 arg(s)")
}

func TestCreateCmd_NoApp(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"create", "staging"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "no app specified")
}

func TestListCmd_NoExtraArgs(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"list", "extra-arg"})

	err := root.Cmd.Execute()
	require.Error(t, err)
}

func TestGetCmd_RequiresNameOrID(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"get"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "environment name or --id")
}

func TestDeleteCmd_RequiresConfirm(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"delete", "staging", "--app", "x"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "--confirm")
}

func TestUpdateCmd_RequiresAtLeastOneField(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"update", "staging", "--app", "x"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "at least one field")
}

func TestUpdateCmd_NamespaceRequiresCluster(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"update", "staging", "--app", "x", "--namespace", "ns"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "--namespace requires --cluster")
}

func TestUpdateCmd_RequiresNameOrID(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"update"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "environment name or --id")
}

func TestDeleteCmd_RequiresNameOrID(t *testing.T) {
	root := newTestEnvCmd(t)
	root.Cmd.SetArgs([]string{"delete"})

	err := root.Cmd.Execute()
	require.Error(t, err)
	require.ErrorContains(t, err, "environment name or --id")
}

// ---------------------------------------------------------------------------
// addAppFlag isolation: each subcommand gets its own --app flag
// ---------------------------------------------------------------------------

func TestAddAppFlag_Isolation(t *testing.T) {
	root := NewEnvCmd(&factory.Options{})

	// Find create and list subcommands.
	var createCmd, listCmd = findSubCmd(root, "create"), findSubCmd(root, "list")
	require.NotNil(t, createCmd, "create subcommand not found")
	require.NotNil(t, listCmd, "list subcommand not found")

	// Set --app on create, verify list's --app is independent.
	require.NoError(t, createCmd.Flags().Set("app", "billing-api"))

	listVal, err := listCmd.Flags().GetString("app")
	require.NoError(t, err)
	require.Equal(t, "", listVal, "list --app should be independent of create --app")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func findSubCmd(root *EnvCmd, name string) *cobra.Command {
	for _, c := range root.Cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
