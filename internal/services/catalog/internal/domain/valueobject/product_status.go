package valueobject

import "errors"

// ProductStatus represents the lifecycle state of a Product aggregate.
// Only forward transitions are allowed (draft → active → sold | archived).
type ProductStatus struct {
	value string
}

var (
	// ProductStatusDraft means the product is a draft; not visible to buyers.
	ProductStatusDraft = ProductStatus{"draft"}
	// ProductStatusPublished means the product is live and can be purchased.
	ProductStatusPublished = ProductStatus{"published"}
	// ProductStatusSold means the product has been sold.
	ProductStatusSold = ProductStatus{"sold"}
	// ProductStatusArchived means the product has been archived by the seller.
	ProductStatusArchived = ProductStatus{"archived"}
)

// AllProductStatuses returns the full set of valid status values.
func AllProductStatuses() []ProductStatus {
	return []ProductStatus{
		ProductStatusDraft,
		ProductStatusPublished,
		ProductStatusSold,
		ProductStatusArchived,
	}
}

// String returns the raw string value of the status.
func (s ProductStatus) String() string { return s.value }

// CanTransitionTo reports whether the receiver status may legally transition to target.
func (s ProductStatus) CanTransitionTo(target ProductStatus) bool {
	switch s.value {
	case "draft":
		return target.value == "published"
	case "published":
		return target.value == "sold" || target.value == "archived"
	case "sold", "archived":
		return false // terminal states
	}
	return false
}

// NewProductStatusFromString constructs a ProductStatus from its string value.
func NewProductStatusFromString(value string) (ProductStatus, error) {
	switch value {
	case "draft":
		return ProductStatusDraft, nil
	case "published":
		return ProductStatusPublished, nil
	case "sold":
		return ProductStatusSold, nil
	case "archived":
		return ProductStatusArchived, nil
	}
	return ProductStatus{}, errors.New("invalid product status")
}
