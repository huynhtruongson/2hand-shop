package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ListProductRequestsRequest carries query parameters for the list product requests endpoint.
type ListProductRequestsRequest struct {
	Category   *string  `form:"category"`
	Conditions []string `form:"conditions" binding:"omitempty,dive,oneof=new like_new good fair poor"`
	Statuses   []string `form:"statuses" binding:"omitempty,dive,oneof=pending approved rejected"`
	Page       int      `form:"page,default=1" binding:"min=1"`
	Limit      int      `form:"limit,default=20" binding:"min=1,max=100"`
	Sort       *string  `form:"sort" binding:"omitempty,oneof=created_at -created_at expected_price -expected_price"`
}

// ToListProductRequestsQuery converts the HTTP request DTO to the application-layer query struct.
func (req ListProductRequestsRequest) ToListProductRequestsQuery() query.ListProductRequestsQuery {
	return query.ListProductRequestsQuery{
		Page:       req.Page,
		Limit:      req.Limit,
		Category:   req.Category,
		Conditions: req.Conditions,
		Statuses:   req.Statuses,
		Sort:       req.Sort,
	}
}

// ProductRequestSummary is the HTTP response DTO for a single product request.
type ProductRequestSummary struct {
	ID                string                  `json:"id"`
	SellerID          string                  `json:"seller_id"`
	CategoryID        string                  `json:"category_id"`
	Title             string                  `json:"title"`
	Description       string                  `json:"description"`
	Brand             *string                 `json:"brand,omitempty"`
	ExpectedPrice     customtypes.Price       `json:"expected_price"`
	Currency          string                  `json:"currency"`
	Condition         string                  `json:"condition"`
	Images            customtypes.Attachments `json:"images"`
	ContactInfo       string                  `json:"contact_info"`
	Status            string                  `json:"status"`
	AdminRejectReason *string                 `json:"admin_reject_reason,omitempty"`
	AdminNote         *string                 `json:"admin_note,omitempty"`
	CreatedAt         string                  `json:"created_at"`
	UpdatedAt         string                  `json:"updated_at"`
}

// ToProductRequestsDTO converts domain aggregates into HTTP response DTOs.
func ToProductRequestsDTO(domainPRs []*aggregate.ProductRequest) []ProductRequestSummary {
	results := make([]ProductRequestSummary, len(domainPRs))
	for i, pr := range domainPRs {
		results[i] = ProductRequestSummary{
			ID:                pr.ID(),
			SellerID:          pr.SellerID(),
			CategoryID:        pr.CategoryID(),
			Title:             pr.Title(),
			Description:       pr.Description(),
			Brand:             pr.Brand(),
			ExpectedPrice:     pr.ExpectedPrice(),
			Currency:          pr.Currency().String(),
			Condition:         pr.Condition().String(),
			Images:            pr.Images(),
			ContactInfo:       pr.ContactInfo(),
			Status:            pr.Status().String(),
			AdminRejectReason: pr.AdminRejectReason(),
			AdminNote:         pr.AdminNote(),
			CreatedAt:         pr.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         pr.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return results
}
