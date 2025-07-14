package akashicpay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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
	CryptoUSDT CryptoCurrency = "USDT"
	CryptoTRX  CryptoCurrency = "TRX"
	CryptoETH  CryptoCurrency = "ETH"
	CryptoSEP  CryptoCurrency = "SEP"
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

func (ap *AkashicPay) Payout(recipientId string, to string, amount string, network NetworkSymbol, token TokenSymbol) (string, error) {
	if recipientId == "" {
		return "", errors.New("recipientId may not be zero-valued")
	}
	if to == "" {
		return "", errors.New("to may not be zero-valued")
	}
	if amount == "" {
		return "", errors.New("amount may not be zero-valued")
	}
	if network == "" {
		return "", errors.New("network may not be zero-valued")
	}
	ToAddress := to
	InitiatedToNonL2 := ""
	IsL2 := false

	DecimalAmount, err := ConvertToSmallestUnit(amount, network, token)
	if err != nil {
		return "", err
	}

	L2Lookup, err := ap.LookForL2Address(to, network)

	if err != nil {
		return "", err
	}

	InputIsL1, err := regexp.MatchString(NetworkDictionary[network].AddressRegex, to)
	if err != nil {
		return "", err
	}
	InputIsL2, err := regexp.MatchString(L2RegexWithOptionalPrefix, to)
	if err != nil {
		return "", err
	}

	if InputIsL1 {
		if L2Lookup.L2Address != "" {
			ToAddress = L2Lookup.L2Address
			InitiatedToNonL2 = to
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
		InitiatedToNonL2 = to
		IsL2 = true
	}

	// L2
	if IsL2 {
		signedL2Tx, err := L2Transaction(ap.Env, ap.Otk, network, DecimalAmount, ToAddress, token, InitiatedToNonL2, recipientId, ap.IsFxBp)
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
		ToAddress:             to,
		Amount:                amount,
		NetworkSymbol:         network,
		TokenSymbol:           token,
		Identifier:            recipientId,
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

	SignedTxn, err := signTransaction(res.PreparedTxn, ap.Otk)

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

func (ap *AkashicPay) GetDepositUrl(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, redirectUrl string) (string, error) {
	return ap.getDepositUrlFunc(identifier, referenceId, receiveCurrencies, redirectUrl, "", "", 0)
}

func (ap *AkashicPay) GetDepositUrlWithRequestedValue(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, redirectUrl string, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (string, error) {
	if referenceId == "" {
		return "", errors.New("referenceId may not be zero-valued")
	}
	if requestedCurrency == "" {
		return "", errors.New("requestedCurrency may not be zero-valued")
	}
	if requestedAmount == "" {
		return "", errors.New("requestedAmount may not be zero-valued")
	}
	return ap.getDepositUrlFunc(identifier, referenceId, receiveCurrencies, redirectUrl, requestedCurrency, requestedAmount, markupPercentage)
}

func (ap *AkashicPay) GetDepositAddress(network NetworkSymbol, identifier string, referenceId string) (IDepositAddress, error) {
	return ap.getDepositAddressFunc(network, identifier, referenceId, "", "", "", 0)
}

func (ap *AkashicPay) GetDepositAddressWithRequestedValue(network NetworkSymbol, identifier string, referenceId string, requestedCurrency Currency, requestedAmount string, token TokenSymbol, markupPercentage float64) (IDepositAddress, error) {
	if referenceId == "" {
		return IDepositAddress{}, errors.New("referenceId may not be zero-valued")
	}
	if requestedCurrency == "" {
		return IDepositAddress{}, errors.New("requestedCurrency may not be zero-valued")
	}
	if requestedAmount == "" {
		return IDepositAddress{}, errors.New("requestedAmount may not be zero-valued")
	}
	return ap.getDepositAddressFunc(network, identifier, referenceId, token, requestedCurrency, requestedAmount, markupPercentage)
}

func (ap *AkashicPay) GetExchangeRates(requestedCurrency Currency) (IGetExchangeRatesResult, error) {
	if requestedCurrency == "" {
		return IGetExchangeRatesResult{}, errors.New("requestedCurrency may not be zero-valued")
	}
	return getExchangeRates(ap.AkashicUrl, requestedCurrency)
}

func (ap *AkashicPay) LookForL2Address(aliasOrL1OrL2Address string, network NetworkSymbol) (ILookForL2AddressResponse, error) {
	if aliasOrL1OrL2Address == "" {
		return ILookForL2AddressResponse{}, errors.New("aliasOrL1OrL2Address may not be zero-valued")
	}
	return getL2Lookup(ap.AkashicUrl, aliasOrL1OrL2Address, network)
}

func (ap *AkashicPay) GetTransfers(getTransactionParams IGetTransactions) ([]ITransaction, error) {
	validLimits := map[int]bool{0: true, 10: true, 25: true, 50: true, 100: true}
	if !validLimits[getTransactionParams.Limit] {
		return nil, errors.New("limit must be one of 10, 25, 50, or 100")
	}
	return getTransfers(ap.AkashicUrl, ap.Otk.Identity, getTransactionParams)
}

func (ap *AkashicPay) GetTransactionDetails(l2Hash string) (ITransaction, error) {
	if l2Hash == "" {
		return ITransaction{}, errors.New("l2Hash may not be zero-valued")
	}
	return getTransactionDetails(ap.AkashicUrl, l2Hash)
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

func (ap *AkashicPay) getDepositUrlFunc(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, redirectUrl string, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (string, error) {
	if identifier == "" {
		return "", errors.New("identifier may not be zero-valued")
	}
	keys, err := getKeysByOwnerAndIdentifier(ap.AkashicUrl, ap.Otk.Identity, identifier)
	if err != nil {
		return "", err
	}
	supportedCurrencies, err := getSupportedCurrencies(ap.AkashicUrl)
	if err != nil {
		return "", err
	}

	// get networkSymbols that are owned
	existingSymbols := make(map[NetworkSymbol]bool)
	for _, key := range keys {
		existingSymbols[key.CoinSymbol] = true
	}

	// get deposit addresses for unique symbols that are not already owned
	for _, networkSymbols := range supportedCurrencies {
		for _, networkSymbol := range networkSymbols {
			if _, exists := existingSymbols[networkSymbol]; !exists {
				_, err := ap.GetDepositAddress(networkSymbol, identifier, "")
				existingSymbols[networkSymbol] = true
				if err != nil {
					return "", err
				}
			}
		}
	}

	if referenceId != "" {
		payload := ICreateDepositOrder{
			Identity:    ap.Otk.Identity,
			ReferenceId: referenceId,
			Identifier:  identifier,
			Expires:     time.Now().UnixMilli() + 60*1000,
		}
		if markupPercentage != 0 {
			payload.MarkupPercentage = markupPercentage
		}
		if requestedAmount != "" && requestedCurrency != "" {
			payload.RequestedValue = &IRequestedValue{
				Amount:   requestedAmount,
				Currency: requestedCurrency,
			}
		}
		signature, err := signData(payload, ap.Otk.privateKey)
		if err != nil {
			return "", err
		}
		payload.Signature = signature
		// create a deposit order
		_, err = createDepositOrder(ap.AkashicUrl, payload)
		if err != nil {
			return "", err
		}
	}
	params := url.Values{}
	params.Set("identity", ap.Otk.Identity)
	params.Set("identifier", identifier)
	if referenceId != "" {
		params.Set("referenceId", referenceId)
	}
	if len(receiveCurrencies) > 0 {
		params.Set("receiveCurrencies", strings.Join(cryptoCurrencySliceToStringSlice(receiveCurrencies), ","))
	}
	if redirectUrl != "" {
		params.Set("redirectUrl", base64.RawURLEncoding.EncodeToString([]byte(redirectUrl)))
	}
	return fmt.Sprintf("%v/sdk/deposit?%v", ap.AkashicPayUrl, params.Encode()), nil
}

func (ap *AkashicPay) getDepositAddressFunc(network NetworkSymbol, identifier string, referenceId string, token TokenSymbol, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (IDepositAddress, error) {
	// Check environment and network compatibility
	if (ap.Env == Development && (network == Ethereum_Mainnet || network == Tron)) ||
		(ap.Env == Production && (network == Ethereum_Sepolia || network == Tron_Shasta)) {
		return IDepositAddress{}, NewAkashicError(AkashicErrorCodeNetworkEnvironmentMismatch, "")
	}
	if identifier == "" {
		return IDepositAddress{}, errors.New("identifier may not be zero-valued")
	}
	if network == "" {
		return IDepositAddress{}, errors.New("network may not be zero-valued")
	}
	response, err := getByOwnerAndIdentifier(ap.AkashicUrl, network, identifier, ap.Otk.Identity)
	if err != nil {
		return IDepositAddress{}, err
	}
	if response.Address != "" {
		if response.UnassignedLedgerId != "" {
			tx, err := Assign(ap.Env, ap.Otk, response.UnassignedLedgerId, identifier)
			if err != nil {
				return IDepositAddress{}, err
			}

			acRes, err := Post[ActiveLedgerResponse[IKeyCreationResponse, any]](ap.TargetNode.Node, tx)
			if err != nil {
				return IDepositAddress{}, err
			}

			acErr := checkForAkashicChainError(acRes)
			if acErr != nil {
				return IDepositAddress{}, acErr
			}
		}

		if referenceId != "" {
			depositOrder, err := ap.createDepositPayloadAndOrder(referenceId, identifier, response.Address, network, token, requestedCurrency, requestedAmount, markupPercentage)
			if err != nil {
				return IDepositAddress{}, err
			}
			return IDepositAddress{
				Address:           response.Address,
				Identifier:        identifier,
				ReferenceId:       referenceId,
				RequestedAmount:   depositOrder.RequestedAmount,
				RequestedCurrency: depositOrder.RequestedCurrency,
				Network:           network,
				Token:             depositOrder.TokenSymbol,
				ExchangeRate:      depositOrder.ExchangeRate,
				Amount:            depositOrder.Amount,
				Expires:           depositOrder.Expires,
				MarkupPercentage:  depositOrder.MarkupPercentage,
			}, nil
		}
		return IDepositAddress{
			Address:    response.Address,
			Identifier: identifier,
		}, nil
	}

	// If no address found, create a new key
	tx, err := KeyCreateTransaction(ap.Env, network, ap.Otk)
	if err != nil {
		return IDepositAddress{}, err
	}

	createKeyRes, err := Post[ActiveLedgerResponse[IKeyCreationResponse, any]](ap.TargetNode.Node, tx)
	if err != nil {
		return IDepositAddress{}, err
	}
	acErr := checkForAkashicChainError(createKeyRes)
	if acErr != nil {
		return IDepositAddress{}, acErr
	}
	newKey := createKeyRes.Responses[0]

	diffConTx, err := differentialConsensusTransaction(ap.Env, ap.Otk, newKey, identifier)
	if err != nil {
		return IDepositAddress{}, err
	}

	diffConTxResp, err := Post[ActiveLedgerResponse[any, any]](ap.TargetNode.Node, diffConTx)
	if err != nil {
		return IDepositAddress{}, err
	}
	acError := checkForAkashicChainError(diffConTxResp)
	if acError != nil {
		return IDepositAddress{}, acError
	}

	// If referenceId is provided, create a deposit order with new key address
	if referenceId != "" {
		depositOrder, err := ap.createDepositPayloadAndOrder(referenceId, identifier, newKey.Address, network, token, requestedCurrency, requestedAmount, markupPercentage)
		if err != nil {
			return IDepositAddress{}, err
		}
		return IDepositAddress{
			Address:           newKey.Address,
			Identifier:        identifier,
			ReferenceId:       referenceId,
			RequestedAmount:   depositOrder.RequestedAmount,
			RequestedCurrency: depositOrder.RequestedCurrency,
			Network:           network,
			Token:             depositOrder.TokenSymbol,
			ExchangeRate:      depositOrder.ExchangeRate,
			Amount:            depositOrder.Amount,
			Expires:           depositOrder.Expires,
			MarkupPercentage:  depositOrder.MarkupPercentage,
		}, nil
	}
	return IDepositAddress{
		Address:    newKey.Address,
		Identifier: identifier,
	}, nil
}

func (ap *AkashicPay) createDepositPayloadAndOrder(referenceId string, identifier string, address string, network NetworkSymbol, tokenSymbol TokenSymbol, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (ICreateDepositOrderResponse, error) {
	payload := ICreateDepositOrder{
		Identity:    ap.Otk.Identity,
		Expires:     time.Now().UnixMilli() + 60*1000,
		ReferenceId: referenceId,
		Identifier:  identifier,
		ToAddress:   address,
		CoinSymbol:  network,
		TokenSymbol: tokenSymbol,
	}

	if markupPercentage != 0 {
		payload.MarkupPercentage = markupPercentage
	}

	if requestedCurrency != "" && requestedAmount != "" {
		payload.RequestedValue = &IRequestedValue{
			Currency: requestedCurrency,
			Amount:   requestedAmount,
		}
	}

	signature, err := signData(payload, ap.Otk.privateKey)
	if err != nil {
		return ICreateDepositOrderResponse{}, err
	}
	createOrderPayload := payload
	createOrderPayload.Signature = signature
	return createDepositOrder(ap.AkashicUrl, createOrderPayload)
}
