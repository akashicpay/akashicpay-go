package akashicpay

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Akashic Requests

type feeDelegationStrategy string

const (
	feeDelegationNone      feeDelegationStrategy = "None"
	ffeeDelegationDelegate feeDelegationStrategy = "Delegate"
)

type prepareTxnDto struct {
	ToAddress             string                `json:"toAddress"`
	NetworkSymbol         NetworkSymbol         `json:"coinSymbol"`
	Amount                string                `json:"amount"`
	TokenSymbol           TokenSymbol           `json:"tokenSymbol,omitempty"`
	Identity              string                `json:"identity"`
	ReferenceId           string                `json:"referenceId"`
	FeeDelegationStrategy feeDelegationStrategy `json:"feeDelegationStrategy"`
}
type prepareL2TxnDto struct {
	SignedTx acTransaction `json:"signedTx"`
}

// AkashicScan Responses
type isBpResponse struct {
	IsBp   bool `json:"isBp"`
	IsFxBp bool `json:"isFxBp"`
}

type iOwnerBalancesResponse struct {
	CoinSymbol  NetworkSymbol `json:"coinSymbol"`
	Amount      string        `json:"balance"`
	TokenSymbol TokenSymbol   `json:"tokenSymbol"`
}

type iOwnerDetailsResponse struct {
	TotalBalances          []iOwnerBalancesResponse `json:"totalBalances"`
	PendingDepositBalances []iOwnerBalancesResponse `json:"pendingDepositBalances"`
	PendingSendBalances    []iOwnerBalancesResponse `json:"pendingSendBalances"`
	OwnerIdentity          string                   `json:"ownerIdentity"`
	TransactionCount       int                      `json:"transactionCount"`
	IsFxBp                 bool                     `json:"isFxBp"`
}

type ILookForL2AddressResponse struct {
	L2Address string `json:"l2Address,omitempty"`
	Alias     string `json:"alias,omitempty"`
}

type iPrepareL1TxnResponse struct {
	FromAddress  string        `json:"fromAddress"`
	PreparedTxn  acTransaction `json:"preparedTxn"`
	DelegatedFee string        `json:"delegatedFee"`
}

type iPrepareL2TxnResponse struct {
	PreparedTxn acTransaction `json:"preparedTxn"`
}

type IGetExchangeRatesResult map[string]string

type transactionsResponse struct {
	Transactions []ITransaction `json:"transactions"`
}

type l2HashTransactionResponse struct {
	Transaction ITransaction `json:"transaction"`
}

// Akashic Endpoints
const (
	isBpEndpoint                = "/v0/owner/is-bp"
	prepareTxEndpoint           = "/v0/l1-txn-orchestrator/prepare-withdrawal"
	l2LookupEndpoint            = "/v0/nft/look-for-l2-address"
	ownerTransactionEndpoint    = "/v0/owner/transactions"
	ownerBalanceEndpoint        = "/v0/owner/details"
	transactionsDetailsEndpoint = "/v0/transactions/transfer"
	identifierLookupEndpoint    = "/v0/key/bp-deposit-key"
	allKeysOfIdentifierEndpoint = "/v0/key/all-bp-deposit-keys"
	supportedCurrenciesEndpoint = "/v1/config/supported-currencies"
	createDepositOrderEndpoint  = "/v0/deposit-request"
	ownerKeysEndpoint           = "/v0/owner/keys?address"
	prepareL2TxnEndpoint        = "/v0/l2-txn-orchestrator/prepare-l2-withdrawal"
	exchangeRatesEndpoint       = "/v0/exchange-rate"
)

func getIsBp(baseUrl string, l2Address string) (isBpResponse, error) {
	url := fmt.Sprintf("%v%v?address=%v",
		baseUrl,
		isBpEndpoint,
		l2Address,
	)
	isBp, err := get[isBpResponse](url)
	return isBp, err
}

func getBalance(baseUrl string, l2Address string) (iOwnerDetailsResponse, error) {
	url := fmt.Sprintf("%v%v?address=%v",
		baseUrl,
		ownerBalanceEndpoint,
		l2Address,
	)
	ownerDetails, err := get[iOwnerDetailsResponse](url)
	return ownerDetails, err
}

func getL2Lookup(baseUrl string, l2AddressOrAlias string, network NetworkSymbol) (ILookForL2AddressResponse, error) {
	url := fmt.Sprintf("%v%v?to=%v",
		baseUrl,
		l2LookupEndpoint,
		l2AddressOrAlias,
	)
	if network != "" {
		url = fmt.Sprintf("%v&coinSymbol=%v",
			url,
			network,
		)
	}
	l2Lookup, err := get[ILookForL2AddressResponse](url)
	return l2Lookup, err
}

func prepareL1Txn(baseUrl string, payload prepareTxnDto) (iPrepareL1TxnResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, prepareTxEndpoint)
	return post[iPrepareL1TxnResponse](url, payload)
}

func prepareL2Txn(baseUrl string, payload prepareL2TxnDto) (iPrepareL2TxnResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, prepareL2TxnEndpoint)
	return post[iPrepareL2TxnResponse](url, payload)
}

func getSupportedCurrencies(baseUrl string) (map[CryptoCurrency][]NetworkSymbol, error) {
	url := fmt.Sprintf("%v%v",
		baseUrl,
		supportedCurrenciesEndpoint,
	)
	supportedCurrencies, err := get[map[CryptoCurrency][]NetworkSymbol](url)

	return supportedCurrencies, err
}

func getExchangeRates(baseUrl string, requestedCurrency Currency) (IGetExchangeRatesResult, error) {
	url := fmt.Sprintf("%v%v/%v",
		baseUrl,
		exchangeRatesEndpoint,
		requestedCurrency,
	)
	exchangeRates, err := get[IGetExchangeRatesResult](url)

	return exchangeRates, err
}

func getTransfers(baseUrl string, identity string, params IGetTransactions) ([]ITransaction, error) {
	query := getTransfersQueryParams(params, identity)
	url := baseUrl + ownerTransactionEndpoint + "?" + query
	resp, err := get[transactionsResponse](url)

	transactions := resp.Transactions
	return transactions, err
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
		values = append(values, "startDate="+params.StartDate.UTC().Format(time.RFC3339))
	}
	if !params.EndDate.IsZero() {
		values = append(values, "endDate="+params.EndDate.UTC().Format(time.RFC3339))
	}
	if params.Layer != "" {
		values = append(values, "layer="+string(params.Layer))
	}
	if params.Status != "" {
		values = append(values, "status="+string(params.Status))
	}
	if params.TransactionType != "" {
		values = append(values, "transactionType="+string(params.TransactionType))
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
		transactionsDetailsEndpoint,
		l2Hash,
	)
	response, err := get[l2HashTransactionResponse](url)

	t := response.Transaction
	return t, err
}

func getByOwnerAndIdentifier(baseUrl string, coinSymbol NetworkSymbol, identifier string, identity string) (iGetByOwnerAndIdentifierResponse, error) {
	url := fmt.Sprintf("%v%v?identity=%v&identifier=%v&coinSymbol=%v&usePreSeed=true",
		baseUrl,
		identifierLookupEndpoint,
		identity,
		identifier,
		coinSymbol,
	)
	return get[iGetByOwnerAndIdentifierResponse](url)
}

func createDepositOrder(baseUrl string, payload iCreateDepositOrder) (iCreateDepositOrderResponse, error) {
	url := fmt.Sprintf("%v%v", baseUrl, createDepositOrderEndpoint)
	return post[iCreateDepositOrderResponse](url, payload)
}

/**
 * Get all keys by BP and identifier
 */
func getKeysByOwnerAndIdentifier(
	baseUrl string,
	identity string,
	identifier string,
) ([]iKeyByOwnerAndIdentifierResponse, error) {
	Params := url.Values{}
	Params.Set("identity", identity)
	Params.Set("identifier", identifier)
	url := fmt.Sprintf("%v%v?%v", baseUrl, allKeysOfIdentifierEndpoint, Params.Encode())
	return get[[]iKeyByOwnerAndIdentifierResponse](url)
}
