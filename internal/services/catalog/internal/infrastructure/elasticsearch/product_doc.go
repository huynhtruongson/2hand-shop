package elasticsearch

// productDoc is the Elasticsearch document representation of a Product.
// type productDoc struct {
// 	ID           string                  `json:"id"`
// 	CategoryID   string                  `json:"category_id"`
// 	CategoryName string                  `json:"category_name"`
// 	Title        string                  `json:"title"`
// 	Description  string                  `json:"description"`
// 	Brand        *string                 `json:"brand,omitempty"`
// 	Price        string                  `json:"price"`
// 	Currency     string                  `json:"currency"`
// 	Condition    string                  `json:"condition"`
// 	Status       string                  `json:"status"`
// 	Images       customtypes.Attachments `json:"images"`
// 	CreatedAt    time.Time               `json:"created_at"`
// 	UpdatedAt    time.Time               `json:"updated_at"`
// 	DeletedAt    *time.Time              `json:"deleted_at,omitempty"`
// }

// // toProductDoc maps a domain Product aggregate + category name to an Elasticsearch document.
// func toProductDoc(p *aggregate.Product, categoryName string) productDoc {
// 	images := make([]string, 0, len(p.Images()))
// 	for _, a := range p.Images() {
// 		// The S3 URL is constructed from the bucket + key; Attachment stores key separately.
// 		images = append(images, a.Key)
// 	}

// 	return productDoc{
// 		ID:           p.ID(),
// 		CategoryID:   p.CategoryID(),
// 		CategoryName: categoryName,
// 		Title:        p.Title(),
// 		Description:  p.Description(),
// 		Brand:        p.Brand(),
// 		Price:        p.Price().String(),
// 		Currency:     p.Currency().String(),
// 		Condition:    p.Condition().String(),
// 		Status:       p.Status().String(),
// 		Images:       images,
// 		CreatedAt:    p.CreatedAt(),
// 		UpdatedAt:    p.UpdatedAt(),
// 	}
// }
