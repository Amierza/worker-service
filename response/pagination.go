package response

import "gorm.io/gorm"

type (
	PaginationRequest struct {
		Page    int `form:"page"`
		PerPage int `form:"per_page"`
	}

	PaginationResponse struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		MaxPage int64 `json:"max_page"`
		Count   int64 `json:"count"`
	}
)

func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

func (pr *PaginationResponse) GetLimit() int {
	return pr.PerPage
}

func (pr *PaginationResponse) GetPage() int {
	return pr.Page
}

func Paginate(page, perPage int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * perPage
		return db.Offset(offset).Limit(perPage)
	}
}
