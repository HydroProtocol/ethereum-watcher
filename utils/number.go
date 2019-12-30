package utils

import (
	"github.com/shopspring/decimal"
	"math/big"
)

// To Decimal
func StringToDecimal(str string) decimal.Decimal {
	if len(str) >= 2 && str[:2] == "0x" {
		b := new(big.Int)
		b.SetString(str[2:], 16)
		d := decimal.NewFromBigInt(b, 0)
		return d
	} else {
		v, err := decimal.NewFromString(str)
		if err != nil {
			panic(err)
		}
		return v
	}
}
