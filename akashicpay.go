package akashicpay

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
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
	Tron_Shasta      NetworkSymbol = "TRX-SHASTA"
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
func (ap *AkashicPay) Payout(RecipientId string, To string, Amount string, Network NetworkSymbol, Token TokenSymbol) (string, error) {
	ToAddress := To
	InitiatedToNonL2 := ""
	IsL2 := false

	DecimalAmount, err := ConvertToSmallestUnit(Amount, Network, Token)
	if err != nil {
		return "", err
	}

	L2Lookup, err := ap.LookForL2Address(To, Network)

	if err != nil {
		return "", err
	}

	InputIsL1, err := regexp.MatchString(NetworkDictionary[Network].AddressRegex, To)
	if err != nil {
		return "", err
	}
	InputIsL2, err := regexp.MatchString(L2RegexWithOptionalPrefix, To)
	if err != nil {
		return "", err
	}

	if InputIsL1 {
		if L2Lookup.L2Address != "" {
			ToAddress = L2Lookup.L2Address
			InitiatedToNonL2 = To
			IsL2 = true
		}
	} else if InputIsL2 {
		if L2Lookup.L2Address == "" {
			return "", NewAkashicError(AkashicErrorCodeL2AddressNotFound, "")
		}
		IsL2 = true
	} else {
		// Must be alias
		if L2Lookup.L2Address == "" {
			return "", NewAkashicError(AkashicErrorCodeL2AddressNotFound, "")
		}
		ToAddress = L2Lookup.L2Address
		InitiatedToNonL2 = To
		IsL2 = true
	}

	// L2
	if IsL2 {
		signedL2Tx, err := L2Transaction(ap.Env, ap.Otk, Network, DecimalAmount, ToAddress, Token, InitiatedToNonL2, RecipientId, ap.IsFxBp)
		if err != nil {
			return "", err
		}

		//If FX, double-sign on BE
		if ap.IsFxBp {
			res, err := prepareL2Txn(ap.AkashicUrl, PrepareL2TxnDto{SignedTx: signedL2Tx})
			if err != nil {
				return "", err
			}
			signedL2Tx = res.PreparedTxn
		}

		acRes, err := Post[ActiveLedgerResponse[any, any]](ap.TargetNode.Node, signedL2Tx)
		if err != nil {
			return "", err
		}
		acErr := checkForAkashicChainError(acRes)
		if acErr != nil {
			return "", acErr
		}

		return PrefixWithAS(acRes.Umid)
	}

	// L1

	Payload := PrepareTxnDto{
		ToAddress:             To,
		Amount:                Amount,
		NetworkSymbol:         Network,
		TokenSymbol:           Token,
		Identifier:            RecipientId,
		Identity:              ap.Otk.Identity,
		FeeDelegationStrategy: FeeDelegationDelegate,
	}
	res, err := prepareL1Txn(ap.AkashicUrl, Payload)

	if err != nil {
		if strings.Contains(err.Error(), "exceeds total savings") {
			return "", NewAkashicError(AkashicErrorCodeSavingsExceeded, "")
		}
		return "", err
	}

	SignedTxn, err := SignTransaction(res.PreparedTxn, ap.Otk)

	if err != nil {
		return "", err
	}

	acRes, err := Post[ActiveLedgerResponse[any, any]](ap.TargetNode.Node, SignedTxn)

	if err != nil {
		return "", err
	}
	acErr := checkForAkashicChainError(acRes)
	if acErr != nil {
		return "", acErr
	}

	return PrefixWithAS(acRes.Umid)
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

func (ap *AkashicPay) GetTransfers(getTransactionParams IGetTransactions) ([]ITransaction, error) {
	return getTransfers(ap.AkashicUrl, ap.Otk.Identity, getTransactionParams)
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
