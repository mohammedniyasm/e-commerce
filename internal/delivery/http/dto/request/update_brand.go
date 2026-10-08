package request

type UpdateBrandRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}
