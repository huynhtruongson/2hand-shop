package elasticsearch

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// productDoc is the Elasticsearch document representation of a Product.
type productDoc struct {
	ID           string    `json:"id"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Brand        *string   `json:"brand,omitempty"`
	Price        string    `json:"price"`
	Currency     string    `json:"currency"`
	Condition    string    `json:"condition"`
	Status       string    `json:"status"`
	Images       []string  `json:"images"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// toProductDoc maps a domain Product aggregate + category name to an Elasticsearch document.
func toProductDoc(p *aggregate.Product, categoryName string) productDoc {
	images := make([]string, 0, len(p.Images()))
	for _, a := range p.Images() {
		// The S3 URL is constructed from the bucket + key; Attachment stores key separately.
		images = append(images, a.Key)
	}

	return productDoc{
		ID:           p.ID(),
		CategoryID:   p.CategoryID(),
		CategoryName: categoryName,
		Title:        p.Title(),
		Description:  p.Description(),
		Brand:        p.Brand(),
		Price:        p.Price().String(),
		Currency:     p.Currency().String(),
		Condition:    p.Condition().String(),
		Status:       p.Status().String(),
		Images:       images,
		CreatedAt:    p.CreatedAt(),
		UpdatedAt:    p.UpdatedAt(),
	}
}

// attachmentsFromStrings reconstructs a customtypes.Attachments from a list of keys.
// Note: Other fields (Name, Size, ContentType, Type) are zeroed since they are not stored in ES.
// Use this only when reading from ES for display purposes.
func attachmentsFromStrings(keys []string) customtypes.Attachments {
	as := make(customtypes.Attachments, len(keys))
	for i, key := range keys {
		as[i] = customtypes.Attachment{Key: key}
	}
	return as
}
