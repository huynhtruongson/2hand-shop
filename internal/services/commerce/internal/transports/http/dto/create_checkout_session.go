package dto

// CreateCheckoutSessionRequestDTO is the HTTP request body for POST /checkout/sessions.
type CreateCheckoutSessionRequestDTO struct {
	SuccessURL string `json:"success_url" binding:"required,url"`
	CancelURL  string `json:"cancel_url"  binding:"required,url"`
}

// CreateCheckoutSessionResponseDTO is the HTTP response body for POST /checkout/sessions.
type CreateCheckoutSessionResponseDTO struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}
