package resolve

import (
	"testing"

	"github.com/stretchr/testify/require"

	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

func TestVariableType(t *testing.T) {
	tests := []struct {
		input   string
		want    variablev1.VariableType
		wantErr string
	}{
		{input: "string", want: variablev1.VariableType_VARIABLE_TYPE_STRING},
		{input: "str", want: variablev1.VariableType_VARIABLE_TYPE_STRING},
		{input: "STRING", want: variablev1.VariableType_VARIABLE_TYPE_STRING},
		{input: "number", want: variablev1.VariableType_VARIABLE_TYPE_NUMBER},
		{input: "num", want: variablev1.VariableType_VARIABLE_TYPE_NUMBER},
		{input: "boolean", want: variablev1.VariableType_VARIABLE_TYPE_BOOLEAN},
		{input: "bool", want: variablev1.VariableType_VARIABLE_TYPE_BOOLEAN},
		{input: "complex", want: variablev1.VariableType_VARIABLE_TYPE_COMPLEX},
		{input: "Complex", want: variablev1.VariableType_VARIABLE_TYPE_COMPLEX},
		{input: "invalid", wantErr: "unsupported variable type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := VariableType(tt.input)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVariableFilter(t *testing.T) {
	tests := []struct {
		name  string
		appID string
		envID string
		want  string
	}{
		{
			name: "global (empty)",
			want: "",
		},
		{
			name:  "app only",
			appID: "app-123",
			want:  "field['application_id'] = 'app-123'",
		},
		{
			name:  "app and env",
			appID: "app-123",
			envID: "env-456",
			want:  "field['application_id'] = 'app-123' AND field['environment_id'] = 'env-456'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VariableFilter(tt.appID, tt.envID)
			require.Equal(t, tt.want, got)
		})
	}
}
