package admin

import "testing"

func TestIsLookupAccountID(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"28003410-ad59-4d40-9f60-b4837b6a8e5c", true},
		{"42acd776-e502-d006-558e-5577f39a0f85", true},
		{"42acd776e502d006558e5577f39a0f85", false},
		{"42acd776", false},
		{"42acd776...9a0f85", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := isLookupAccountID(tc.query); got != tc.want {
			t.Fatalf("isLookupAccountID(%q) = %v, want %v", tc.query, got, tc.want)
		}
	}
}
