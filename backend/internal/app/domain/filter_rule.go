package domain

import "strings"

const FilterRuleSourceAll = "*"

type FilterType string

const (
	FilterTypeLanguage    FilterType = "language"
	FilterTypeDisplayName FilterType = "display_name"
	FilterTypeUsername    FilterType = "username"
)

type FilterRule struct {
	ID     int64      `json:"id"`
	Source string     `json:"source"`
	Type   FilterType `json:"type"`
	Value  string     `json:"value"`
}

func (r *FilterRule) Matches(source, username, displayName, language string) bool {
	if r.Source != FilterRuleSourceAll && r.Source != source {
		return false
	}

	ruleValue := strings.ToLower(r.Value)

	switch r.Type {
	case FilterTypeUsername:
		return strings.ToLower(username) == ruleValue

	case FilterTypeDisplayName:
		return strings.Contains(strings.ToLower(displayName), ruleValue)

	case FilterTypeLanguage:
		return language != "" && strings.ToLower(language) == ruleValue
	}

	return false
}
