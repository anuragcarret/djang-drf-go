package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Paginator defines the interface for all pagination classes
type Paginator interface {
	PaginateQueryset(request *http.Request, count int) (limit, offset int, err error)
	GetPaginatedResponse(results interface{}, count int, request *http.Request) map[string]interface{}
}

// PageNumberPagination implements page-based pagination
type PageNumberPagination struct {
	PageSize         int    // Default page size
	PageQueryParam   string // Query parameter name for page number
	PageSizeParam    string // Query parameter name for custom page size
	MaxPageSize      int    // Maximum allowed page size
	LastPageStrings  []string
	InvalidPageValue string
}

func NewPageNumberPagination() *PageNumberPagination {
	return &PageNumberPagination{
		PageSize:         10,
		PageQueryParam:   "page",
		PageSizeParam:    "page_size",
		MaxPageSize:      100,
		LastPageStrings:  []string{"last"},
		InvalidPageValue: "invalid",
	}
}

func (p *PageNumberPagination) PaginateQueryset(request *http.Request, count int) (limit, offset int, err error) {
	// Get page number
	pageStr := request.URL.Query().Get(p.PageQueryParam)
	page := 1
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return 0, 0, fmt.Errorf("invalid page number")
		}
	}

	// Get page size
	pageSize := p.PageSize
	if p.PageSizeParam != "" {
		if sizeStr := request.URL.Query().Get(p.PageSizeParam); sizeStr != "" {
			customSize, err := strconv.Atoi(sizeStr)
			if err == nil && customSize > 0 {
				pageSize = customSize
				// Enforce max page size
				if p.MaxPageSize > 0 && pageSize > p.MaxPageSize {
					pageSize = p.MaxPageSize
				}
			}
		}
	}

	offset = (page - 1) * pageSize
	limit = pageSize

	// Check if page is within range
	if offset >= count && count > 0 {
		return 0, 0, fmt.Errorf("invalid page")
	}

	return limit, offset, nil
}

func (p *PageNumberPagination) GetPaginatedResponse(results interface{}, count int, request *http.Request) map[string]interface{} {
	limit, offset, _ := p.PaginateQueryset(request, count)

	currentPage := (offset / limit) + 1
	totalPages := (count + limit - 1) / limit

	// Build next/previous links
	baseURL := request.URL
	query := baseURL.Query()

	var nextLink, prevLink *string

	if currentPage < totalPages {
		query.Set(p.PageQueryParam, strconv.Itoa(currentPage+1))
		next := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		nextLink = &next
	}

	if currentPage > 1 {
		query.Set(p.PageQueryParam, strconv.Itoa(currentPage-1))
		prev := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		prevLink = &prev
	}

	return map[string]interface{}{
		"count":    count,
		"next":     nextLink,
		"previous": prevLink,
		"results":  results,
	}
}

// LimitOffsetPagination implements limit/offset-based pagination
type LimitOffsetPagination struct {
	DefaultLimit     int
	LimitQueryParam  string
	OffsetQueryParam string
	MaxLimit         int
}

func NewLimitOffsetPagination() *LimitOffsetPagination {
	return &LimitOffsetPagination{
		DefaultLimit:     10,
		LimitQueryParam:  "limit",
		OffsetQueryParam: "offset",
		MaxLimit:         100,
	}
}

func (p *LimitOffsetPagination) PaginateQueryset(request *http.Request, count int) (limit, offset int, err error) {
	// Get limit
	limit = p.DefaultLimit
	if limitStr := request.URL.Query().Get(p.LimitQueryParam); limitStr != "" {
		customLimit, err := strconv.Atoi(limitStr)
		if err == nil && customLimit > 0 {
			limit = customLimit
			// Enforce max limit
			if p.MaxLimit > 0 && limit > p.MaxLimit {
				limit = p.MaxLimit
			}
		}
	}

	// Get offset
	offset = 0
	if offsetStr := request.URL.Query().Get(p.OffsetQueryParam); offsetStr != "" {
		customOffset, err := strconv.Atoi(offsetStr)
		if err == nil && customOffset >= 0 {
			offset = customOffset
		}
	}

	return limit, offset, nil
}

func (p *LimitOffsetPagination) GetPaginatedResponse(results interface{}, count int, request *http.Request) map[string]interface{} {
	limit, offset, _ := p.PaginateQueryset(request, count)

	// Build next/previous links
	baseURL := request.URL
	query := baseURL.Query()

	var nextLink, prevLink *string

	if offset+limit < count {
		query.Set(p.LimitQueryParam, strconv.Itoa(limit))
		query.Set(p.OffsetQueryParam, strconv.Itoa(offset+limit))
		next := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		nextLink = &next
	}

	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		query.Set(p.LimitQueryParam, strconv.Itoa(limit))
		query.Set(p.OffsetQueryParam, strconv.Itoa(prevOffset))
		prev := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		prevLink = &prev
	}

	return map[string]interface{}{
		"count":    count,
		"next":     nextLink,
		"previous": prevLink,
		"results":  results,
	}
}

// CursorPagination implements cursor-based stable pagination
type CursorPagination struct {
	PageSize         int
	CursorQueryParam string
	Ordering         string // Field to order by (required)
}

type cursorData struct {
	Offset     int         `json:"o"`
	OrderValue interface{} `json:"v"`
}

func NewCursorPagination(ordering string) *CursorPagination {
	return &CursorPagination{
		PageSize:         10,
		CursorQueryParam: "cursor",
		Ordering:         ordering,
	}
}

func (p *CursorPagination) PaginateQueryset(request *http.Request, count int) (limit, offset int, err error) {
	if p.Ordering == "" {
		return 0, 0, fmt.Errorf("cursor pagination requires an ordering field")
	}

	limit = p.PageSize
	offset = 0

	// Decode cursor if present
	cursorStr := request.URL.Query().Get(p.CursorQueryParam)
	if cursorStr != "" {
		decoded, err := base64.URLEncoding.DecodeString(cursorStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid cursor")
		}

		var cursor cursorData
		if err := json.Unmarshal(decoded, &cursor); err != nil {
			return 0, 0, fmt.Errorf("invalid cursor format")
		}

		offset = cursor.Offset
	}

	return limit, offset, nil
}

func (p *CursorPagination) GetPaginatedResponse(results interface{}, count int, request *http.Request) map[string]interface{} {
	limit, offset, _ := p.PaginateQueryset(request, count)

	// Build next/previous links
	baseURL := request.URL
	query := baseURL.Query()

	var nextLink, prevLink *string

	if offset+limit < count {
		// Encode next cursor
		cursor := cursorData{
			Offset: offset + limit,
		}
		cursorJSON, _ := json.Marshal(cursor)
		cursorStr := base64.URLEncoding.EncodeToString(cursorJSON)

		query.Set(p.CursorQueryParam, cursorStr)
		next := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		nextLink = &next
	}

	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}

		cursor := cursorData{
			Offset: prevOffset,
		}
		cursorJSON, _ := json.Marshal(cursor)
		cursorStr := base64.URLEncoding.EncodeToString(cursorJSON)

		query.Set(p.CursorQueryParam, cursorStr)
		prev := baseURL.Scheme + "://" + baseURL.Host + baseURL.Path + "?" + query.Encode()
		prevLink = &prev
	}

	return map[string]interface{}{
		"count":    nil, // Cursor pagination typically doesn't return count
		"next":     nextLink,
		"previous": prevLink,
		"results":  results,
	}
}

// Helper to apply pagination to URL values
func ApplyPagination(values url.Values, paginator Paginator) url.Values {
	// This would be called by views to apply pagination settings
	return values
}
