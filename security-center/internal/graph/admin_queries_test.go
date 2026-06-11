package graph

import "testing"

func TestNormalizeDeviceQuery(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"42acd776e502d006558e5577f39a0f85", "42acd776e502d006558e5577f39a0f85"},
		{"42ACD776E502D006558E5577F39A0F85", "42acd776e502d006558e5577f39a0f85"},
		{"42acd776...9a0f85", "42acd776"},
		{"42acd776…9a0f85", "42acd776"},
		{"  42acd776  ", "42acd776"},
		{"...", ""},
	}

	for _, tc := range tests {
		if got := normalizeDeviceQuery(tc.in); got != tc.want {
			t.Fatalf("normalizeDeviceQuery(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
