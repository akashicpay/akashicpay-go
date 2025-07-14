package akashicpay

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

type TransactionsResponse struct {
	Transactions []ITransaction `json:"transactions"`
}

type L2HashTransactionResponse struct {
	Transaction ITransaction `json:"transaction"`
}

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

func getTransfers(baseUrl string, identity string, params IGetTransactions) ([]ITransaction, error) {
	query := getTransfersQueryParams(params, identity)
	url := baseUrl + OwnerTransactionEndpoint + "?" + query
	resp, err := Get[TransactionsResponse](url)
	return resp.Transactions, err
}

func getTransfersQueryParams(params IGetTransactions, identity string) string {
	values := make([]string, 0)
	if params.Page != 0 {
		values = append(values, "page="+strconv.Itoa(params.Page))
	}
	if params.Limit != 0 {
		values = append(values, "limit="+strconv.Itoa(params.Limit))
	}
	if !params.StartDate.IsZero() {
		values = append(values, "startDate="+params.StartDate.Format(time.RFC3339))
	}
	if !params.EndDate.IsZero() {
		values = append(values, "endDate="+params.EndDate.Format(time.RFC3339))
	}
	if params.Layer != "" {
		values = append(values, "layer="+string(params.Layer))
	}
	if params.Status != "" {
		values = append(values, "status="+string(params.Status))
	}
	if params.Type != "" {
		values = append(values, "type="+string(params.Type))
	}
	if params.HideSmallTransactions {
		values = append(values, "hideSmallTransactions=true")
	}
	values = append(values, "identity="+identity)
	return joinQueryParams(values)
}

func joinQueryParams(params []string) string {
	return strings.Join(params, "&")
}

func getTransactionDetails(baseUrl string, l2Hash string) (ITransaction, error) {
	url := fmt.Sprintf("%v%v?l2Hash=%v",
		baseUrl,
		TransactionsDetailsEndpoint,
		l2Hash,
	)
	response, err := Get[L2HashTransactionResponse](url)
	return response.Transaction, err
}

func getByOwnerAndIdentifier(baseUrl string, coinSymbol NetworkSymbol, identifier string, identity string) (IGetByOwnerAndIdentifierResponse, error) {
	url := fmt.Sprintf("%v%v?identity=%v&identifier=%v&coinSymbol=%v&usePreSeed=true",
		baseUrl,
		IdentifierLookupEndpoint,
		identity,
		identifier,
		coinSymbol,
	)
	return Get[IGetByOwnerAndIdentifierResponse](url)
}

func createDepositOrder(baseUrl string, payload ICreateDepositOrder) (ICreateDepositOrderResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, CreateDepositOrderEndpoint)
	return Post[ICreateDepositOrderResponse](url, payload)
}

/**
 * Get all keys by BP and identifier
 */
func getKeysByOwnerAndIdentifier(
	baseUrl string,
	identifier string,
) (KeyResponseWrapper, error) {
	url := fmt.Sprintf("%v%v?identifier=%v", baseUrl, AllKeysOfIdentifierEndpoint, identifier)
	return Get[KeyResponseWrapper](url)
}
