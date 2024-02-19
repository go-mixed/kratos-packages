package db

type Pagination struct {
	Page   int   `json:"page"`             // 页码
	Limit  int   `json:"limit"`            // 每页条数
	Total  int64 `json:"total"`            // 总数据条数
	Offset int   `json:"offset,omitempty"` // 自定义偏移值
}

func DefaultPagination() *Pagination {
	return &Pagination{
		Page:  1,
		Limit: 10,
	}
}

func NewPagination(page, limit int) *Pagination {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	return &Pagination{
		Page:  page,
		Limit: limit,
	}
}

// GetOffset 获取分页偏移值
func (p *Pagination) GetOffset() int {
	if p.Offset > 0 {
		return p.Offset
	}
	offset := 0
	if p.Page > 0 {
		offset = (p.Page - 1) * p.Limit
	}
	return offset
}

// TotalPage 获取总页数
func (p *Pagination) TotalPage() int64 {
	if p.Total == 0 || p.Limit == 0 {
		return 0
	}
	totalPage := p.Total / int64(p.Limit)
	if p.Total%int64(p.Limit) > 0 {
		totalPage = totalPage + 1
	}
	return totalPage
}
