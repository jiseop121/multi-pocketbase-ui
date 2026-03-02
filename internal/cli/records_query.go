package cli

import (
	"strings"
)

type RecordsQueryState struct {
	Collection string
	Page       int
	PerPage    int
	Sort       string
	Filter     string
	Fields     []string
}

func (s RecordsQueryState) QueryParams() map[string]string {
	query := map[string]string{}
	if strings.TrimSpace(s.Sort) != "" {
		query["sort"] = s.Sort
	}
	if strings.TrimSpace(s.Filter) != "" {
		query["filter"] = s.Filter
	}
	if s.Page > 0 {
		query["page"] = intToString(s.Page)
	}
	if s.PerPage > 0 {
		query["perPage"] = intToString(s.PerPage)
	}
	if len(s.Fields) > 0 {
		query["fields"] = strings.Join(s.Fields, ",")
	}
	return query
}

func normalizeColumns(cols []string) []string {
	set := map[string]struct{}{}
	cleaned := make([]string, 0, len(cols))
	for _, col := range cols {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}
		if _, ok := set[col]; ok {
			continue
		}
		set[col] = struct{}{}
		cleaned = append(cleaned, col)
	}
	return cleaned
}
