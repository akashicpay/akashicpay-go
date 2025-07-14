package akashicpay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Environment string
type NetworkSymbol string
type TokenSymbol string
type CryptoCurrency string

const (
	Development Environment = "Development"
	Production  Environment = "Production"
)

const (
	Tron             NetworkSymbol = "TRX"
	Tron_Shasta      NetworkSymbol = "TRX_SHASTA"
	Ethereum_Mainnet NetworkSymbol = "ETH"
	Ethereum_Sepolia NetworkSymbol = "SEP"
)

const (
	USDT TokenSymbol = "USDT"
)

// Currencies supported by AkashicPay, includes native coins and tokens
const (
	USDT_C CryptoCurrency = "USDT"
	TRX    CryptoCurrency = "TRX"
	ETH    CryptoCurrency = "ETH"
	SEP    CryptoCurrency = "SEP"
)

type AkashicPay struct {
	IsFxBp        bool
	Otk           Otk
	TargetNode    ACNode
	AkashicUrl    string
	AkashicPayUrl string
	Env           Environment
	ApiSecret     string
}

type Balance struct {
	NetworkSymbol NetworkSymbol
	TokenSymbol   TokenSymbol
	Balance       string
}

// Construct and initialize a new AkashicPay instance. Returns a pointer to an
// AkashicPay instance
func NewAkashicPay(privateKey string, identity string, env Environment, apiSecret string) (*AkashicPay, error) {
	otk, err := reconstructOtkFromPrivateKey(privateKey, identity)
	if err != nil {
		return nil, err
	}

	fastestNode, err := chooseBestACNode(env)
	if err != nil {
		return nil, err
	}

	urls := getUrls(env)

	isBp, err := getIsBp(urls.AkashicUrl, identity)
	if err != nil {
		return nil, err
	}
	if !isBp.IsBp {
		return nil, NewAkashicError(AkashicErrorCodeIsNotBp, "")
	}

	return &AkashicPay{
		IsFxBp:        isBp.IsFxBp,
		TargetNode:    fastestNode,
		AkashicUrl:    urls.AkashicUrl,
		AkashicPayUrl: urls.AkashicPayUrl,
		Env:           env,
		ApiSecret:     apiSecret,
		Otk:           otk,
	}, nil
}

// Get total balances, divided by Network and Token
func (ap *AkashicPay) GetBalance() ([]Balance, error) {
	ownerDetails, err := getBalance(ap.AkashicUrl, ap.Otk.Identity)

	if err != nil {
		return nil, err
	}
	balances := make([]Balance, len(ownerDetails.TotalBalances))
	for i, v := range ownerDetails.TotalBalances {
		balances[i] = Balance{
			NetworkSymbol: v.CoinSymbol,
			TokenSymbol:   v.TokenSymbol,
			Balance:       v.Amount,
		}
	}
	return balances, nil
}

// TODO: Implement below++
func (ap *AkashicPay) Payout() {

}

func (ap *AkashicPay) GetDepositAddress() {

}

func (ap *AkashicPay) GetDepositUrl() {

}

func (ap *AkashicPay) GetExchangeRates(requestedCurrency Currency) (IGetExchangeRatesResult, error) {
	return getExchangeRates(ap.AkashicUrl, requestedCurrency)
}

func (ap *AkashicPay) LookForL2Address(aliasOrL1OrL2Address string, network NetworkSymbol) (ILookForL2AddressResponse, error) {
	return getL2Lookup(ap.AkashicUrl, aliasOrL1OrL2Address, network)
}

func (ap *AkashicPay) GetTransfers() {

}

func (ap *AkashicPay) GetTransactionDetails(l2Hash string) {

}

// Get the currently supported currencies in AkashicPay
// Returns a map from currencies to a list of networks supported for that currency
func (ap *AkashicPay) GetSupportedCurrencies() (map[CryptoCurrency][]NetworkSymbol, error) {
	return getSupportedCurrencies(ap.AkashicUrl)
}

func chooseBestACNode(env Environment) (ACNode, error) {
	var nodes map[string]ACNode
	if env == Production {
		nodes = ACNodes
	} else {
		nodes = ACDevNodes
	}

	type Result struct {
		Node   ACNode
		Status int
		Error  error
	}

	type AcResponse struct {
		Status int `json:"status"`
	}

	// Use a goroutine to concurrently check node health
	results := make(chan Result, len(nodes))
	for _, node := range nodes {
		go func(node ACNode) {
			resp, err := http.Get(node.Node + "a/status")
			if err != nil {
				results <- Result{ACNode{}, 0, err}
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				results <- Result{ACNode{}, 0, err}
				return
			}

			var nodeData AcResponse
			err = json.Unmarshal(body, &nodeData)
			if err != nil {
				results <- Result{ACNode{}, 0, err}
				return
			}

			results <- Result{node, nodeData.Status, nil}
		}(node)
	}

	for i := range results {
		// When first 4 comes, return that node
		if i.Status == 4 {
			return i.Node, nil
		}
	}

	// if we don't get any 4s we return an error
	return ACNode{}, errors.New("no healthy AC node")
}
