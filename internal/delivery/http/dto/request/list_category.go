package request

type ListCategoryRequest struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}
