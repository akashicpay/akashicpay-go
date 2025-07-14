package akashicpay

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	AkashicBaseUrlDev string = "https://api.testnet.akashicscan.com/api"
	AkashicBaseUrl    string = "https://api.akashicscan.com/api"
)

const (
	AkashicPayBaseUrlDev string = "https://testnet.akashicpay.com"
	AkashicPayBaseUrl    string = "https://www.akashicpay.com"
)

const L2RegexWithOptionalPrefix = `^(AS)?[A-Fa-f\d]{64}$`

type SdkUrls struct {
	AkashicUrl    string
	AkashicPayUrl string
}

func getUrls(environment Environment) SdkUrls {
	if environment == Development {
		return SdkUrls{
			AkashicUrl:    AkashicBaseUrlDev,
			AkashicPayUrl: AkashicPayBaseUrlDev,
		}
	}

	return SdkUrls{
		AkashicUrl:    AkashicBaseUrl,
		AkashicPayUrl: AkashicPayBaseUrl,
	}
}

// Ensures a UMID/L2-address has exactly one AS prefix. i.e. it is idempotent.
// returns error if the umid argument isn't a valid UMID/L2-address with or
// without the prefix
func PrefixWithAS(umid string) (string, error) {
	r, err := regexp.Compile(L2RegexWithOptionalPrefix)
	matches := r.FindAllStringSubmatch(umid, -1)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("%s does not match regex with or without prefix", umid)
	}
	if matches[0][1] != "" {
		return umid, nil
	}
	return "AS" + umid, nil
}

// Ensures a UMID/L2-address doesn't have an AS prefix. i.e. it is idempotent.
// returns error if the umid argument isn't a valid UMID/L2-address with or
// without the prefix
func RemoveASPrefix(umid string) (string, error) {
	r, err := regexp.Compile(L2RegexWithOptionalPrefix)
	matches := r.FindAllStringSubmatch(umid, -1)
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("%s does not match regex with or without prefix", umid)
	}
	if matches[0][1] != "" {
		return strings.TrimPrefix(umid, "AS"), nil
	}
	return umid, nil
}
