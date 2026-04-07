package valueobject

import "errors"

// OrderStatus represents the lifecycle state of an Order aggregate.
// Transitions: pending → confirmed → shipped → delivered | cancelled
type OrderStatus struct {
	value string
}

var (
	// OrderStatusPending means the order has been placed but payment is not yet confirmed.
	OrderStatusPending = OrderStatus{"pending"}
	// OrderStatusConfirmed means payment has been confirmed and the order is being processed.
	OrderStatusConfirmed = OrderStatus{"confirmed"}
	// OrderStatusShipped means the order has been shipped to the buyer.
	OrderStatusShipped = OrderStatus{"shipped"}
	// OrderStatusDelivered means the order has been delivered to the buyer.
	OrderStatusDelivered = OrderStatus{"delivered"}
	// OrderStatusCancelled means the order has been cancelled.
	OrderStatusCancelled = OrderStatus{"cancelled"}
)

// AllOrderStatuses returns the full set of valid status values.
func AllOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusConfirmed,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
	}
}

// String returns the raw string value of the status.
func (s OrderStatus) String() string { return s.value }

// CanTransitionTo reports whether the receiver status may legally transition to target.
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	switch s.value {
	case "pending":
		return target.value == "confirmed" || target.value == "cancelled"
	case "confirmed":
		return target.value == "shipped" || target.value == "cancelled"
	case "shipped":
		return target.value == "delivered" || target.value == "cancelled"
	case "delivered", "cancelled":
		return false // terminal states
	}
	return false
}

// NewOrderStatusFromString constructs an OrderStatus from its string value.
func NewOrderStatusFromString(value string) (OrderStatus, error) {
	switch value {
	case "pending":
		return OrderStatusPending, nil
	case "confirmed":
		return OrderStatusConfirmed, nil
	case "shipped":
		return OrderStatusShipped, nil
	case "delivered":
		return OrderStatusDelivered, nil
	case "cancelled":
		return OrderStatusCancelled, nil
	}
	return OrderStatus{}, errors.New("invalid order status")
}