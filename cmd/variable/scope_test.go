package variable

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveScope(t *testing.T) {
	tests := []struct {
		name       string
		global     bool
		env        string
		appArg     string
		contextApp string
		wantScope  resolvedScope
		wantErr    string
	}{
		{
			name:       "context app with env",
			contextApp: "billing-api",
			env:        "staging",
			wantScope:  resolvedScope{Scope: scopeAppEnv, App: "billing-api", Env: "staging"},
		},
		{
			name:       "context app without env",
			contextApp: "billing-api",
			wantScope:  resolvedScope{Scope: scopeApp, App: "billing-api"},
		},
		{
			name:       "context app with global",
			contextApp: "billing-api",
			global:     true,
			wantScope:  resolvedScope{Scope: scopeGlobal},
		},
		{
			name:      "arg app with env",
			appArg:    "my-api",
			env:       "staging",
			wantScope: resolvedScope{Scope: scopeAppEnv, App: "my-api", Env: "staging"},
		},
		{
			name:      "arg app without env",
			appArg:    "my-api",
			wantScope: resolvedScope{Scope: scopeApp, App: "my-api"},
		},
		{
			name:       "arg app overrides context",
			appArg:     "my-api",
			contextApp: "billing-api",
			wantScope:  resolvedScope{Scope: scopeApp, App: "my-api"},
		},
		{
			name:      "global scope",
			global:    true,
			wantScope: resolvedScope{Scope: scopeGlobal},
		},
		{
			name:    "no app no global",
			wantErr: "no app specified",
		},
		{
			name:    "global with env",
			global:  true,
			env:     "staging",
			wantErr: "--global and --env are mutually exclusive",
		},
		{
			name:    "global with app arg",
			global:  true,
			appArg:  "my-api",
			wantErr: "--global and app argument are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveScope(tt.global, tt.env, tt.appArg, tt.contextApp)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantScope, got)
		})
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantApp string
		wantKV  []string
		wantErr string
	}{
		{
			name:   "kv only",
			args:   []string{"KEY=val"},
			wantKV: []string{"KEY=val"},
		},
		{
			name:    "app and kv",
			args:    []string{"my-api", "KEY=val"},
			wantApp: "my-api",
			wantKV:  []string{"KEY=val"},
		},
		{
			name:    "app and multiple kv",
			args:    []string{"my-api", "A=1", "B=2"},
			wantApp: "my-api",
			wantKV:  []string{"A=1", "B=2"},
		},
		{
			name:    "two bare args",
			args:    []string{"my-api", "other"},
			wantErr: "unexpected argument",
		},
		{
			name:    "kv before app",
			args:    []string{"A=1", "my-api"},
			wantApp: "my-api",
			wantKV:  []string{"A=1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, kv, err := splitArgs(tt.args)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantApp, app)
			require.Equal(t, tt.wantKV, kv)
		})
	}
}

func TestSplitArgsWithKey(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantApp string
		wantKey string
		wantErr string
	}{
		{
			name:    "key only",
			args:    []string{"IMAGE_TAG"},
			wantKey: "IMAGE_TAG",
		},
		{
			name:    "app and key",
			args:    []string{"my-api", "IMAGE_TAG"},
			wantApp: "my-api",
			wantKey: "IMAGE_TAG",
		},
		{
			name:    "too many args",
			args:    []string{"a", "b", "c"},
			wantErr: "expected 1-2 arguments",
		},
		{
			name:    "kv rejected",
			args:    []string{"KEY=val"},
			wantErr: "unexpected KEY=VALUE argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, key, err := splitArgsWithKey(tt.args)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantApp, app)
			require.Equal(t, tt.wantKey, key)
		})
	}
}

func TestParseKV(t *testing.T) {
	tests := []struct {
		input   string
		wantK   string
		wantV   string
		wantErr string
	}{
		{input: "KEY=value", wantK: "KEY", wantV: "value"},
		{input: "K=v=w", wantK: "K", wantV: "v=w"},
		{input: "KEY=", wantK: "KEY", wantV: ""},
		{input: "=value", wantErr: "invalid variable format"},
		{input: "noequals", wantErr: "invalid variable format"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			k, v, err := parseKV(tt.input)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantK, k)
			require.Equal(t, tt.wantV, v)
		})
	}
}
