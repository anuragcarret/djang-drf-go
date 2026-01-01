package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// Page represents a paginated result
type Page struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  interface{} `json:"results"`
}

// PageNumberPagination handles numeric page-based pagination
type PageNumberPagination struct {
	PageSize int
}

func NewPageNumberPagination(pageSize int) *PageNumberPagination {
	return &PageNumberPagination{PageSize: pageSize}
}

func (p *PageNumberPagination) GetPaginatedResponse(data interface{}, count int, r *http.Request) Page {
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}

	totalPages := int(math.Ceil(float64(count) / float64(p.PageSize)))

	var next, prev *string
	if page < totalPages {
		n := r.URL.String() + "?page=" + strconv.Itoa(page+1)
		next = &n
	}
	if page > 1 {
		pr := r.URL.String() + "?page=" + strconv.Itoa(page-1)
		prev = &pr
	}

	return Page{
		Count:    count,
		Next:     next,
		Previous: prev,
		Results:  data,
	}
}
