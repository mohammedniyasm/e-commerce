package request

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateProductImageRequestVariantID(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		wantVariant *int64
	}{
		{name: "omitted keeps variant unchanged", payload: `{}`},
		{name: "null selects product level", payload: `{"variant_id":null}`, wantVariant: int64Pointer(0)},
		{name: "numeric id selects variant", payload: `{"variant_id":12}`, wantVariant: int64Pointer(12)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request UpdateProductImageRequest
			require.NoError(t, json.Unmarshal([]byte(tt.payload), &request))
			if tt.wantVariant == nil {
				require.Nil(t, request.VariantID)
				return
			}
			require.NotNil(t, request.VariantID)
			require.Equal(t, *tt.wantVariant, *request.VariantID)
		})
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
