package request

type CreateBrandRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}