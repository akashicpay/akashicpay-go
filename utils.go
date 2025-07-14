package akashicpay

import (
	"fmt"
	"regexp"
)

const (
	akashicBaseUrlDev string = "https://api.testnet.akashicscan.com/api"
	akashicBaseUrl    string = "https://api.akashicscan.com/api"
)

const (
	akashicPayBaseUrlDev string = "https://testnet.akashicpay.com"
	akashicPayBaseUrl    string = "https://www.akashicpay.com"
)

const l2RegexWithOptionalPrefix = `^(AS)?[A-Fa-f\d]{64}$`

type sdkUrls struct {
	AkashicUrl    string
	AkashicPayUrl string
}

func getUrls(environment Environment) sdkUrls {
	if environment == Development {
		return sdkUrls{
			AkashicUrl:    akashicBaseUrlDev,
			AkashicPayUrl: akashicPayBaseUrlDev,
		}
	}

	return sdkUrls{
		AkashicUrl:    akashicBaseUrl,
		AkashicPayUrl: akashicPayBaseUrl,
	}
}

// Ensures a UMID/L2-address has exactly one AS prefix. i.e. it is idempotent.
// returns error if the umid argument isn't a valid UMID/L2-address with or
// without the prefix
func prefixWithAS(umid string) (string, error) {
	r, err := regexp.Compile(l2RegexWithOptionalPrefix)
	matches := r.FindAllStringSubmatch(umid, -1)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("%s does not match regex with or without prefix", umid)
	}
	if matches[0][1] != "" {
		return umid, nil
	}
	return "AS" + umid, nil
}

// Converts a slice of Currency to a slice of strings.
func cryptoCurrencySliceToStringSlice(currencies []CryptoCurrency) []string {
	result := make([]string, len(currencies))
	for i, c := range currencies {
		result[i] = string(c) // explicit conversion
	}
	return result
}

// Normalize token symbols (map TETHER to USDT)
func normalizeTokenSymbol(symbol TokenSymbol) TokenSymbol {
	if symbol == tether {
		return USDT
	}
	return symbol
}

// Normalize token (map USDT to TETHER for Tron Shasta network)
func normalizeTokenInput(network NetworkSymbol, token TokenSymbol) TokenSymbol {
	if network == Tron_Shasta && token == USDT {
		return tether
	}
	return token
}
