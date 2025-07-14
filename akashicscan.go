package akashicpay

import "fmt"

// Akashic Requests

type FeeDelegationStrategy string

const (
	FeeDelegationNone     FeeDelegationStrategy = "None"
	FeeDelegationDelegate FeeDelegationStrategy = "Delegate"
)

type PrepareTxnDto struct {
	ToAddress             string                `json:"toAddress"`
	NetworkSymbol         NetworkSymbol         `json:"coinSymbol"`
	Amount                string                `json:"amount"`
	TokenSymbol           TokenSymbol           `json:"tokenSymbol,omitempty"`
	Identity              string                `json:"identity"`
	Identifier            string                `json:"identifier"`
	FeeDelegationStrategy FeeDelegationStrategy `json:"feeDelegationStrategy"`
}
type PrepareL2TxnDto struct {
	SignedTx ACTransaction `json:"signedTx"`
}

// AkashicScan Responses
type IsBpResponse struct {
	IsBp   bool `json:"isBp"`
	IsFxBp bool `json:"isFxBp"`
}

type IOwnerBalancesResponse struct {
	CoinSymbol  NetworkSymbol `json:"coinSymbol"`
	Amount      string        `json:"balance"`
	TokenSymbol TokenSymbol   `json:"tokenSymbol"`
}

type IOwnerDetailsResponse struct {
	TotalBalances          []IOwnerBalancesResponse `json:"totalBalances"`
	PendingDepositBalances []IOwnerBalancesResponse `json:"pendingDepositBalances"`
	PendingSendBalances    []IOwnerBalancesResponse `json:"pendingSendBalances"`
	OwnerIdentity          string                   `json:"ownerIdentity"`
	TransactionCount       int                      `json:"transactionCount"`
	IsFxBp                 bool                     `json:"isFxBp"`
}

type ILookForL2AddressResponse struct {
	L2Address string `json:"l2Address,omitempty"`
	Alias     string `json:"alias,omitempty"`
}

type IPrepareL1TxnResponse struct {
	FromAddress  string        `json:"fromAddress"`
	PreparedTxn  ACTransaction `json:"preparedTxn"`
	DelegatedFee string        `json:"delegatedFee"`
}

type IPrepareL2TxnResponse struct {
	PreparedTxn ACTransaction `json:"preparedTxn"`
}

type IGetExchangeRatesResult map[string]string

// Akashic Endpoints
const (
	IsBpEndpoint                = "/v0/owner/is-bp"
	PrepareTxEndpoint           = "/v0/l1-txn-orchestrator/prepare-withdrawal"
	L2LookupEndpoint            = "/v0/nft/look-for-l2-address"
	OwnerTransactionEndpoint    = "/v0/owner/transactions"
	OwnerBalanceEndpoint        = "/v0/owner/details"
	TransactionsDetailsEndpoint = "/v0/transactions/transfer"
	IdentifierLookupEndpoint    = "/v0/key/bp-deposit-key"
	AllKeysOfIdentifierEndpoint = "/v0/key/all-bp-deposit-keys"
	SupportedCurrenciesEndpoint = "/v1/config/supported-currencies"
	CreateDepositOrderEndpoint  = "/v0/deposit-request"
	OwnerKeysEndpoint           = "/v0/owner/keys?address"
	PrepareL2TxnEndpoint        = "/v0/l2-txn-orchestrator/prepare-l2-withdrawal"
	ExchangeRatesEndpoint       = "/v0/exchange-rate"
)

func getIsBp(baseUrl string, l2Address string) (IsBpResponse, error) {
	url := fmt.Sprintf("%v%v?address=%v",
		baseUrl,
		IsBpEndpoint,
		l2Address,
	)
	isBp, err := Get[IsBpResponse](url)
	return isBp, err
}

func getBalance(baseUrl string, l2Address string) (IOwnerDetailsResponse, error) {
	url := fmt.Sprintf("%v%v?address=%v",
		baseUrl,
		OwnerBalanceEndpoint,
		l2Address,
	)
	ownerDetails, err := Get[IOwnerDetailsResponse](url)
	return ownerDetails, err
}

func getL2Lookup(baseUrl string, l2AddressOrAlias string, network NetworkSymbol) (ILookForL2AddressResponse, error) {
	url := fmt.Sprintf("%v%v?to=%v",
		baseUrl,
		L2LookupEndpoint,
		l2AddressOrAlias,
	)
	if network != "" {
		url = fmt.Sprintf("%v&coinSymbol=%v",
			url,
			network,
		)
	}
	l2Lookup, err := Get[ILookForL2AddressResponse](url)
	return l2Lookup, err
}

func prepareL1Txn(baseUrl string, payload PrepareTxnDto) (IPrepareL1TxnResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, PrepareTxEndpoint)
	return Post[IPrepareL1TxnResponse](url, payload)
}

func prepareL2Txn(baseUrl string, payload PrepareL2TxnDto) (IPrepareL2TxnResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, PrepareL2TxnEndpoint)
	return Post[IPrepareL2TxnResponse](url, payload)
}

func getSupportedCurrencies(baseUrl string) (map[CryptoCurrency][]NetworkSymbol, error) {
	url := fmt.Sprintf("%v%v",
		baseUrl,
		SupportedCurrenciesEndpoint,
	)
	supportedCurrencies, err := Get[map[CryptoCurrency][]NetworkSymbol](url)

	return supportedCurrencies, err
}

func getExchangeRates(baseUrl string, requestedCurrency Currency) (IGetExchangeRatesResult, error) {
	url := fmt.Sprintf("%v%v/%v",
		baseUrl,
		ExchangeRatesEndpoint,
		requestedCurrency,
	)
	exchangeRates, err := Get[IGetExchangeRatesResult](url)

	return exchangeRates, err
}
