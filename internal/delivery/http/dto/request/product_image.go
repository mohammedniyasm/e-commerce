package request

import (
	"bytes"
	"encoding/json"
)

type UploadProductImagesRequest struct {
	VariantID string `form:"variant_id"`
}

type UpdateProductImageRequest struct {
	VariantID    *int64 `json:"variant_id"`
	DisplayOrder *int   `json:"display_order"`
	IsPrimary    *bool  `json:"is_primary"`
}

func (r *UpdateProductImageRequest) UnmarshalJSON(data []byte) error {
	type updateProductImageRequest UpdateProductImageRequest

	var decoded updateProductImageRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	*r = UpdateProductImageRequest(decoded)
	if value, exists := fields["variant_id"]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		productLevel := int64(0)
		r.VariantID = &productLevel
	}

	return nil
}
