package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterRule_Matches(t *testing.T) {
	tests := []struct {
		name         string
		rule         FilterRule
		inputSource  string
		inputUser    string
		inputDisplay string
		inputLang    string
		want         bool
	}{
		{
			name:        "Username exact match",
			rule:        FilterRule{Source: "*", Type: FilterTypeUsername, Value: "SpamBot"},
			inputSource: "mastodon",
			inputUser:   "spambot", // test case-insensitive match
			want:        true,
		},
		{
			name:        "Username mismatch",
			rule:        FilterRule{Source: "*", Type: FilterTypeUsername, Value: "SpamBot"},
			inputSource: "mastodon",
			inputUser:   "real_user",
			want:        false,
		},
		{
			name:         "Display name contains",
			rule:         FilterRule{Source: "instagram", Type: FilterTypeDisplayName, Value: "Crypto"},
			inputSource:  "instagram",
			inputDisplay: "Get your cheap CRYPTO here!",
			want:         true,
		},
		{
			name:         "Display name mismatch",
			rule:         FilterRule{Source: "instagram", Type: FilterTypeDisplayName, Value: "Crypto"},
			inputSource:  "instagram",
			inputDisplay: "Hello world",
			want:         false,
		},
		{
			name:        "Source mismatch (rule only for mastodon)",
			rule:        FilterRule{Source: "mastodon", Type: FilterTypeUsername, Value: "troll"},
			inputSource: "instagram",
			inputUser:   "troll",
			want:        false,
		},
		{
			name:        "Language match",
			rule:        FilterRule{Source: "*", Type: FilterTypeLanguage, Value: "ru"},
			inputSource: "mastodon",
			inputLang:   "RU",
			want:        true,
		},
		{
			name:        "Language empty input",
			rule:        FilterRule{Source: "*", Type: FilterTypeLanguage, Value: "en"},
			inputSource: "mastodon",
			inputLang:   "",
			want:        false,
		},
		{
			name:         "No rules match at all",
			rule:         FilterRule{Source: "*", Type: FilterTypeUsername, Value: "non-existent-user"},
			inputSource:  "mastodon",
			inputUser:    "happy_user",
			inputDisplay: "Just a normal post",
			inputLang:    "de",
			want:         false,
		},
		{
			name:        "Unknown filter type should not match",
			rule:        FilterRule{Source: "*", Type: "unknown_type", Value: "some_value"},
			inputSource: "mastodon",
			inputUser:   "some_value",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rule.Matches(tt.inputSource, tt.inputUser, tt.inputDisplay, tt.inputLang)
			assert.Equal(
				t,
				tt.want,
				got,
				"Rule: %v, Input: %s/%s/%s",
				tt.rule,
				tt.inputSource,
				tt.inputUser,
				tt.inputLang,
			)
		})
	}
}
