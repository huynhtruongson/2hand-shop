package valueobject

import "errors"

// PaymentStatus represents the lifecycle state of a Payment aggregate.
// Transitions: pending → confirmed → failed | refunded
type PaymentStatus struct {
	value string
}

var (
	// PaymentStatusPending means the payment intent has been created but not yet confirmed.
	PaymentStatusPending = PaymentStatus{"pending"}
	// PaymentStatusConfirmed means the payment has been successfully processed.
	PaymentStatusConfirmed = PaymentStatus{"confirmed"}
	// PaymentStatusFailed means the payment processing failed.
	PaymentStatusFailed = PaymentStatus{"failed"}
	// PaymentStatusRefunded means the payment has been refunded.
	PaymentStatusRefunded = PaymentStatus{"refunded"}
)

// AllPaymentStatuses returns the full set of valid status values.
func AllPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusPending,
		PaymentStatusConfirmed,
		PaymentStatusFailed,
		PaymentStatusRefunded,
	}
}

// String returns the raw string value of the status.
func (s PaymentStatus) String() string { return s.value }

// CanTransitionTo reports whether the receiver status may legally transition to target.
func (s PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	switch s.value {
	case "pending":
		return target.value == "confirmed" || target.value == "failed"
	case "confirmed":
		return target.value == "refunded"
	case "failed", "refunded":
		return false // terminal states
	}
	return false
}

// NewPaymentStatusFromString constructs a PaymentStatus from its string value.
func NewPaymentStatusFromString(value string) (PaymentStatus, error) {
	switch value {
	case "pending":
		return PaymentStatusPending, nil
	case "confirmed":
		return PaymentStatusConfirmed, nil
	case "failed":
		return PaymentStatusFailed, nil
	case "refunded":
		return PaymentStatusRefunded, nil
	}
	return PaymentStatus{}, errors.New("invalid payment status")
}