package akashicpay

import (
	"errors"
	"math/big"
)

type Currency string

// Fiat- and Crypto-currencies for setting amounts in deposit-orders
const (
	CurrencyUSDT Currency = "USDT"
	CurrencyUSDC Currency = "USDC"
	CurrencyTRX  Currency = "TRX"
	CurrencyETH  Currency = "ETH"
	CurrencyBNB  Currency = "BNB"
	CurrencySOL  Currency = "SOL"

	CurrencyCHF Currency = "CHF"
	CurrencyCNY Currency = "CNY"
	CurrencyEUR Currency = "EUR"
	CurrencyHKD Currency = "HKD"
	CurrencyIDR Currency = "IDR"
	CurrencyJPY Currency = "JPY"
	CurrencyKHR Currency = "KHR"
	CurrencyKRW Currency = "KRW"
	CurrencyMYR Currency = "MYR"
	CurrencyPHP Currency = "PHP"
	CurrencySGD Currency = "SGD"
	CurrencyTHB Currency = "THB"
	CurrencyTWD Currency = "TWD"
	CurrencyUSD Currency = "USD"
	CurrencyVND Currency = "VND"
)

// Method for safe conversion from the human-friendly, divisible units displayed
// in the UI, to the smallest, indivisible coin/token unit.
func convertToSmallestUnit(amount string, coinSymbol NetworkSymbol, tokenSymbol TokenSymbol) (string, error) {
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
	pFloat.SetPrec(100)

	floatAmount.Mul(floatAmount, pFloat)

	smallestUnit, _ := floatAmount.Int(nil)

	return smallestUnit.Text(10), nil
}

func getConversionFactor(coinSymbol NetworkSymbol, tokenSymbol TokenSymbol) (int, error) {
	if tokenSymbol == "" {
		return networkDictionary[coinSymbol].NativeDecimal, nil
	}
	var token *token
	for _, t := range networkDictionary[coinSymbol].Tokens {
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
