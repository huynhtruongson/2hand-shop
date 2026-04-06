package dto

// RejectProductRequestDTO is the request body for rejecting a product request.
type RejectProductRequestDTO struct {
	AdminRejectReason string `json:"admin_reject_reason" binding:"required,max=500"`
}

// RejectProductRequestResponseDTO is returned after a product request is successfully rejected.
type RejectProductRequestResponseDTO struct{}
