package schemas

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (p Page) Offset() int { return (p.Page - 1) * p.PageSize }

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
type List struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func Paginated(data any, p Page, total int64) List {
	return List{Data: data, Pagination: Pagination{Page: p.Page, PageSize: p.PageSize, Total: total}}
}
