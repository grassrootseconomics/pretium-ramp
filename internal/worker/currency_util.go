package worker

import (
	"fmt"
	"math/big"

	"github.com/govalues/decimal"
)

func USDToKES(usdtAmountStr string, contractAddress string, kesRateStr float64) (string, error) {
	var decimals = 6
	if contractAddress == "0x765DE816845861e75A25fCA122bb6898B8B1282a" {
		decimals = 18
	}

	usdtString, err := weiToDecimalString(usdtAmountStr, decimals)
	if err != nil {
		return "", err
	}

	usdt, err := decimal.Parse(usdtString)
	if err != nil {
		return "", err
	}

	rate, err := decimal.NewFromFloat64(kesRateStr)
	if err != nil {
		return "", err
	}

	kes, err := usdt.Mul(rate)
	if err != nil {
		return "", err
	}

	kesRounded := kes.Round(0)
	return kesRounded.String(), nil
}

func weiToDecimalString(raw string, decimals int) (string, error) {
	wei, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", fmt.Errorf("invalid wei: %s", raw)
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	r := new(big.Rat).SetFrac(wei, scale)
	return r.FloatString(decimals), nil
}
