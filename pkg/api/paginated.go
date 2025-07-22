package api

import "fmt"

type Paginated[T any] struct {
	Count   int `json:"count"`
	Results []T `json:"results"`
}

type PaginatedRequestFilter = map[string]any

type PaginatedRequest struct {
	Ordering string                 `json:"ordering"`
	Page     int                    `json:"page"`
	Search   string                 `json:"search"`
	Size     int                    `json:"size"`
	Filter   PaginatedRequestFilter `json:"filter"`
}

var defaultValues = PaginatedRequest{
	Size:     10,
	Page:     1,
	Ordering: "",
	Search:   "",
	Filter:   make(map[string]any),
}

func NewPaginatedRequest(size, page int, ordering, search string, filter PaginatedRequestFilter) PaginatedRequest {
	req := defaultValues

	if size > 0 {
		req.Size = size
	}
	if page > 0 {
		req.Page = page
	}
	if ordering != "" {
		req.Ordering = ordering
	}
	if search != "" {
		req.Search = search
	}
	if filter != nil {
		req.Filter = filter
	}

	return req
}

func NewAllPaginatedRequest(filter PaginatedRequestFilter) PaginatedRequest {
	return NewPaginatedRequest(1000, 0, "", "", filter)
}

func ConvertToQueryParams(req PaginatedRequest) QueryParams {
	queryParams := QueryParams{
		"ordering": req.Ordering,
		"page":     fmt.Sprintf("%d", req.Page),
		"size":     fmt.Sprintf("%d", req.Size),
		"search":   req.Search,
	}

	for key, value := range req.Filter {
		queryParams[key] = value
	}

	return queryParams
}
