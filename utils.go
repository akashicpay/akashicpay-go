package akashicpay

const (
	AkashicBaseUrlDev string = "https://api.testnet.akashicscan.com/api"
	AkashicBaseUrl    string = "https://api.akashicscan.com/api"
)

const (
	AkashicPayBaseUrlDev string = "https://testnet.akashicpay.com"
	AkashicPayBaseUrl    string = "https://www.akashicpay.com"
)

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
