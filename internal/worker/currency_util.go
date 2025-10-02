package worker

import (
	"fmt"
	"math/big"

	"github.com/govalues/decimal"
)

const stablesDecimals = 18

func USDToKES(usdtAmountStr string, kesRateStr string) (string, error) {
	usdtString, err := weiToDecimalString(usdtAmountStr, stablesDecimals)
	if err != nil {
		return "", err
	}

	usdt, err := decimal.Parse(usdtString)
	if err != nil {
		return "", err
	}

	rate, err := decimal.Parse(kesRateStr)
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
