package options

import "testing"

func TestGuessStructNameFromTableName(t *testing.T) {
	testcases := []struct {
		input string
		want  string
	}{
		{"users", "User"},
		{"entry_category_relations", "EntryCategoryRelation"},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.input+" -> "+tc.want, func(t *testing.T) {
			got := guessStructNameFromTable(tc.input)
			if got != tc.want {
				t.Errorf("expected '%s', got '%s'", tc.want, got)
			}
		})
	}
}
