package dto

// GetCheckoutSessionResponseDTO is the HTTP response body for GET /checkout/sessions.
type GetCheckoutSessionResponseDTO struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount_cents"`
	Currency  string `json:"currency"`
}
