package postgressqlx

import "fmt"

type Page struct {
	Limit  int
	Offset int
}

func NewPage(limit, offset, maxLimit int) Page {
	if maxLimit <= 0 {
		maxLimit = 20
	}
	if limit <= 0 || limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return Page{Limit: limit, Offset: offset}
}

func (p Page) SQL() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Offset)
}

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)
