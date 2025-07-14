package akashicpay

import (
	"errors"
	"math/big"
)

// Method for safe conversion from the human-friendly, divisible units displayed
// in the UI, to the smallest, indivisible coin/token unit.
func ConvertToSmallestUnit(amount string, coinSymbol NetworkSymbol, tokenSymbol TokenSymbol) (string, error) {
	conversionFactor, err := getConversionFactor(coinSymbol, tokenSymbol)
	if err != nil {
		return "", err
	}
	var p *big.Int = big.NewInt(0)
	var floatAmount *big.Float = big.NewFloat(0)
	floatAmount.SetPrec(100) // To easily handle millions(!) of ETH

	floatAmount, ok := floatAmount.SetString(amount)
	if !ok {
		return "", errors.New("invalid amount")
	}

	p.Exp(big.NewInt(10), big.NewInt(int64(conversionFactor)), nil)

	var pFloat *big.Float = big.NewFloat(0).SetInt(p)

	floatAmount.Mul(floatAmount, pFloat)

	if !floatAmount.IsInt() {
		return "", errors.New("transaction is too small or has too many decimals")
	}

	return floatAmount.Text('f', 0), nil
}

func getConversionFactor(coinSymbol NetworkSymbol, tokenSymbol TokenSymbol) (int, error) {
	if tokenSymbol == "" {
		return NetworkDictionary[coinSymbol].NativeDecimal, nil
	}
	var token *Token
	for _, t := range NetworkDictionary[coinSymbol].Tokens {
		if t.Symbol == tokenSymbol {
			token = &t
			break
		}
	}
	if token == nil {
		return -1, errors.New("coin not supported")
	}
	return token.Decimal, nil
}
