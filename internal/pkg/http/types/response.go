package types

type HttpResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	Data       any            `json:"data,omitempty"`
	Error      *ErrorResponse `json:"error,omitempty"`
	Pagination *Pagination    `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
