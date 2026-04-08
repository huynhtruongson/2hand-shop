package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// Address represents a physical address for shipping or billing.
// It implements sql.Scanner and driver.Valuer so it can be stored as a
// PostgreSQL JSONB column, and json.Marshaler/json.Unmarshaler for
// serialisation in API responses and RabbitMQ events.
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"` // ISO 3166-1 alpha-2, e.g. "US", "VN"
}

// addressJSON is the internal serialisation struct used by Scan/Value to avoid
// infinite recursion.
type addressJSON struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

func (a Address) toJSON() addressJSON {
	return addressJSON{
		Street:     a.Street,
		City:       a.City,
		State:      a.State,
		PostalCode: a.PostalCode,
		Country:    a.Country,
	}
}

// String returns a human-readable representation of the address.
func (a Address) String() string {
	parts := []string{a.Street, a.City, a.State, a.PostalCode, a.Country}
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, ", ")
}

// Scan implements database/sql.Scanner. It deserialises a JSONB value from
// PostgreSQL back into an Address.
func (a *Address) Scan(src any) error {
	if src == nil {
		*a = Address{}
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Address.Scan: expected []byte, got %T", src)
	}
	var aj addressJSON
	if err := json.Unmarshal(b, &aj); err != nil {
		return fmt.Errorf("Address.Scan: failed to unmarshal JSON: %w", err)
	}
	*a = Address{
		Street:     aj.Street,
		City:       aj.City,
		State:      aj.State,
		PostalCode: aj.PostalCode,
		Country:    aj.Country,
	}
	return nil
}

// Value implements database/sql/driver.Valuer. It serialises the Address
// as a JSONB value for PostgreSQL.
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a.toJSON())
}

// MarshalJSON implements json.Marshaler.
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.toJSON())
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Address) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = Address{}
		return nil
	}
	var aj addressJSON
	if err := json.Unmarshal(data, &aj); err != nil {
		return err
	}
	*a = Address{
		Street:     aj.Street,
		City:       aj.City,
		State:      aj.State,
		PostalCode: aj.PostalCode,
		Country:    aj.Country,
	}
	return nil
}

// compile-time interface assertions
var (
	_ driver.Valuer                  = Address{}
	_ interface{ Scan(any) error }   = (*Address)(nil)
	_ json.Marshaler                 = Address{}
	_ json.Unmarshaler               = (*Address)(nil)
)
