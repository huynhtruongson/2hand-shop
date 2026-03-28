package valueobject

import "errors"

// Condition represents the physical condition of a listed item.
// It is a value object: equality is based on its string value,
// not object identity.
type Condition struct {
	value string
}

var (
	// ConditionNew is a brand-new, never-used item.
	ConditionNew = Condition{"new"}
	// ConditionLikeNew is an item in excellent condition with minimal use.
	ConditionLikeNew = Condition{"like_new"}
	// ConditionGood is a used item with some signs of wear but fully functional.
	ConditionGood = Condition{"good"}
	// ConditionFair is a used item with noticeable wear; functional.
	ConditionFair = Condition{"fair"}
	// ConditionPoor is a heavily used item; functional but cosmetically poor.
	ConditionPoor = Condition{"poor"}
)

// AllConditions returns the full ordered list of valid condition values,
// suitable for dropdown or validation use.
func AllConditions() []Condition {
	return []Condition{
		ConditionNew,
		ConditionLikeNew,
		ConditionGood,
		ConditionFair,
		ConditionPoor,
	}
}

// String returns the raw string value of the condition.
func (c Condition) String() string { return c.value }

// IsValid reports whether c is one of the five defined conditions.
func (c Condition) IsValid() bool {
	switch c.value {
	case "new", "like_new", "good", "fair", "poor":
		return true
	}
	return false
}

// NewConditionFromString constructs a Condition from its string representation.
// It returns an error if the value is not a recognised condition.
func NewConditionFromString(value string) (Condition, error) {
	switch value {
	case "new":
		return ConditionNew, nil
	case "like_new":
		return ConditionLikeNew, nil
	case "good":
		return ConditionGood, nil
	case "fair":
		return ConditionFair, nil
	case "poor":
		return ConditionPoor, nil
	}
	return Condition{}, errors.New("invalid condition")
}
