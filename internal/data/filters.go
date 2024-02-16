package data

import (
	"strings"

	validator "github.com/kvnloughead/greenlight/internal"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// sortColumn returns the column to sort by from the filter's Sort field.
// It panics if the sort key is not in the safelist.
func (f *Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.Trim(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection returns the direction in which the sort should occur.
// Possible return values: "ASC" and "DESC".
func (f *Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	} else {
		return "ASC"
	}
}

// limit returns the max number of items in a page, as specified by the
// `page_size` query parameter.
func (f *Filters) limit() int {
	return f.PageSize
}

// offset returns the number of rows to skip when display paginated data beyond
// page 1. Calculated from query parameters as (page - 1) * page_size.
func (f *Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {

	v.Check(f.Page >= 1, "page", "must be at least 1")
	v.Check(f.Page <= 10_000_000, "page", "must be no more than least 10,000,000")
	v.Check(f.PageSize >= 1, "page_size", "must be at least 1")
	v.Check(f.PageSize <= 100, "page_size", "must be no more than 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sorting key")
}
