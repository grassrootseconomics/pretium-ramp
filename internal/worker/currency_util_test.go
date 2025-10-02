package worker

import (
	"testing"
)

func TestUSDToKES(t *testing.T) {
	tests := []struct {
		name            string
		usdtAmountStr   string
		kesRateStr      float64
		contractAddress string
		expected        string
	}{
		{
			name:            "round_up",
			usdtAmountStr:   "1250000000000000000",
			contractAddress: "0x0000000000000000000000000000000000000000",
			kesRateStr:      128.41,
			expected:        "161",
		},
		{
			name:            "round_down",
			usdtAmountStr:   "9700000000000000000",
			contractAddress: "0x0000000000000000000000000000000000000000",

			kesRateStr: 128.59,
			expected:   "1247",
		},
		{
			name:            "large_precision",
			usdtAmountStr:   "23893068279562552000",
			contractAddress: "0x0000000000000000000000000000000000000000",

			kesRateStr: 128.41,
			expected:   "3068",
		},
		{
			name:            "zero_amount",
			usdtAmountStr:   "0",
			kesRateStr:      128.41,
			contractAddress: "0x0000000000000000000000000000000000000000",
			expected:        "0",
		},
		{
			name:            "large_precision",
			usdtAmountStr:   "136568613084213597875",
			contractAddress: "0x0000000000000000000000000000000000000000",

			kesRateStr: 128.41,
			expected:   "17537",
		},
		{
			name:            "usdc_token",
			contractAddress: "0xcebA9300f2b948710d2653dD7B07f33A8B32118C",
			usdtAmountStr:   "500000",
			kesRateStr:      128.41,
			expected:        "64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := USDToKES(tt.usdtAmountStr, tt.contractAddress, tt.kesRateStr)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("USDToKES(%s, %f) = %s; want %s",
					tt.usdtAmountStr, tt.kesRateStr, result, tt.expected)
			}
		})
	}
}
