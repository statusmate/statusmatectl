package api

import "fmt"

type Paginated[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

type PaginatedRequest struct {
	Ordering string                 `json:"ordering"`
	Page     int                    `json:"page"`
	Search   string                 `json:"search"`
	Size     int                    `json:"size"`
	Filter   map[string]interface{} `json:"filter"`
}

var defaultValues = PaginatedRequest{
	Size:     10,
	Page:     1,
	Ordering: "",
	Search:   "",
	Filter:   make(map[string]interface{}),
}

// NewPaginatedReq создает объект PaginatedRequest с возможностью переопределения значений.
func NewPaginatedRequest(size, page int, ordering, search string, filter map[string]interface{}) PaginatedRequest {
	req := defaultValues

	// Если переданы новые значения, переопределяем их
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

func NewAllPaginatedRequest(filter map[string]interface{}) PaginatedRequest {
	return NewPaginatedRequest(1000, 0, "", "", filter)
}

// ConvertToQueryParams преобразует объект PaginatedRequest в карту query параметров.
func ConvertToQueryParams(req PaginatedRequest) map[string]string {
	queryParams := map[string]string{
		"ordering": req.Ordering,
		"page":     fmt.Sprintf("%d", req.Page),
		"size":     fmt.Sprintf("%d", req.Size),
		"search":   req.Search,
	}

	// Добавляем фильтры, если они есть
	for key, value := range req.Filter {
		queryParams[key] = fmt.Sprintf("%v", value) // Преобразуем значение в строку
	}

	return queryParams
}
