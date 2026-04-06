package dto

// ProductRequestPathID holds the product_request_id path parameter.
type ProductRequestPathID struct {
	ProductRequestID string `uri:"product_request_id" binding:"required"`
}

// AcceptProductRequestResponseDTO is returned after a product request is successfully accepted.
type AcceptProductRequestResponseDTO struct {
	ProductID string `json:"product_id"`
}
