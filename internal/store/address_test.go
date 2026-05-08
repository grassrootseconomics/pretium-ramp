package store

import "testing"

func TestNormalizeWalletAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name:  "empty_returns_nil",
			input: "",
			want:  nil,
		},
		{
			name:  "lowercase_normalized_to_checksum",
			input: "0xceba9300f2b948710d2653dd7b07f33a8b32118c",
			want:  "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
		},
		{
			name:  "uppercase_normalized_to_checksum",
			input: "0xCEBA9300F2B948710D2653DD7B07F33A8B32118C",
			want:  "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
		},
		{
			name:  "already_checksum_unchanged",
			input: "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
			want:  "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
		},
		{
			name:  "missing_0x_prefix_normalized",
			input: "ceba9300f2b948710d2653dd7b07f33a8b32118c",
			want:  "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeWalletAddress(tt.input)
			if got != tt.want {
				t.Errorf("normalizeWalletAddress(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}
