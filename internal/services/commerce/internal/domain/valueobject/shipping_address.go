package valueobject

import "strings"

// ShippingAddress holds the delivery address for an order.
type ShippingAddress struct {
	Street  string
	City    string
	State   string
	Zip     string
	Country string
}

// NewShippingAddress constructs a ShippingAddress.
func NewShippingAddress(street, city, state, zip, country string) *ShippingAddress {
	return &ShippingAddress{
		Street:  strings.TrimSpace(street),
		City:    strings.TrimSpace(city),
		State:   strings.TrimSpace(state),
		Zip:     strings.TrimSpace(zip),
		Country: strings.TrimSpace(country),
	}
}

// IsValid returns true if all required fields are non-empty.
func (a *ShippingAddress) IsValid() bool {
	return a != nil &&
		a.Street != "" &&
		a.City != "" &&
		a.Zip != "" &&
		a.Country != ""
}