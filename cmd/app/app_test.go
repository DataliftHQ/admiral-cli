package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveApp(t *testing.T) {
	tests := []struct {
		name       string
		appArg     string
		contextApp string
		want       string
		wantErr    string
	}{
		{
			name:   "arg provided",
			appArg: "billing-api",
			want:   "billing-api",
		},
		{
			name:       "context provided",
			contextApp: "billing-api",
			want:       "billing-api",
		},
		{
			name:       "arg overrides context",
			appArg:     "my-api",
			contextApp: "billing-api",
			want:       "my-api",
		},
		{
			name:    "neither provided",
			wantErr: "no app specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveApp(tt.appArg, tt.contextApp)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
