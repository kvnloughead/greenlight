package data

import (
	validator "github.com/kvnloughead/greenlight/internal"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, f Filters) {

	v.Check(f.Page >= 1, "page", "must be at least 1")
	v.Check(f.Page <= 10_000_000, "page", "must be no more than least 10,000,000")
	v.Check(f.PageSize >= 1, "page_size", "must be at least 1")
	v.Check(f.PageSize <= 100, "page_size", "must be no more than 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sorting key")
}
