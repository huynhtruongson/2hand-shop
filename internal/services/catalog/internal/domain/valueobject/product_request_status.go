package valueobject

import "errors"

// ProductRequestStatus represents the lifecycle state of a ProductRequest aggregate.
// Only forward transitions are allowed (pending → approved | rejected).
// Once approved or rejected, the request is a dead end — no further transitions are allowed.
type ProductRequestStatus struct {
	value string
}

var (
	// ProductRequestStatusPending means the request is awaiting admin review.
	ProductRequestStatusPending = ProductRequestStatus{"pending"}
	// ProductRequestStatusApproved means the request has been approved by an admin.
	ProductRequestStatusApproved = ProductRequestStatus{"approved"}
	// ProductRequestStatusRejected means the request has been rejected by an admin.
	ProductRequestStatusRejected = ProductRequestStatus{"rejected"}
)

// AllProductRequestStatuses returns the full set of valid status values.
func AllProductRequestStatuses() []ProductRequestStatus {
	return []ProductRequestStatus{
		ProductRequestStatusPending,
		ProductRequestStatusApproved,
		ProductRequestStatusRejected,
	}
}

// String returns the raw string value of the status.
func (s ProductRequestStatus) String() string { return s.value }

// CanTransitionTo reports whether the receiver status may legally transition to target.
func (s ProductRequestStatus) CanTransitionTo(target ProductRequestStatus) bool {
	switch s.value {
	case "pending":
		return target.value == "approved" || target.value == "rejected"
	case "approved", "rejected":
		return false // terminal states — no further transitions allowed
	}
	return false
}

// NewProductRequestStatusFromString constructs a ProductRequestStatus from its string value.
func NewProductRequestStatusFromString(value string) (ProductRequestStatus, error) {
	switch value {
	case "pending":
		return ProductRequestStatusPending, nil
	case "approved":
		return ProductRequestStatusApproved, nil
	case "rejected":
		return ProductRequestStatusRejected, nil
	}
	return ProductRequestStatus{}, errors.New("invalid product request status")
}
