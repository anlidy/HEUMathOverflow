package api

// Pagination describes basic pagination metadata.
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// Response defines a common HTTP JSON envelope used by
// controllers to return data in a consistent shape.
// Code is typically aligned with HTTP status codes for now,
// but can evolve to business error codes later.
type Response struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	// TraceID    string      `json:"trace_id,omitempty"`
}
