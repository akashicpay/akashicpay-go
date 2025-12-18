// Package akashicpay provides functions to easily interact with the
// AkashicChain network
//
// This includes making payouts, creating wallets for deposits, and querying transaction-details
package akashicpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

// Network supported by AkashicPay, test- and mainnets
const (
	Tron                        NetworkSymbol = "TRX"
	Tron_Shasta                 NetworkSymbol = "TRX-SHASTA"
	Ethereum_Mainnet            NetworkSymbol = "ETH"
	Ethereum_Sepolia            NetworkSymbol = "SEP"
	Binance_Smart_Chain_Mainnet NetworkSymbol = "BNB"
	Binance_Smart_Chain_Testnet NetworkSymbol = "tBNB"
	Solana                      NetworkSymbol = "SOL"
	Solana_Devnet               NetworkSymbol = "SOLDEV"
)

// Tokens supported by AkashicPay
const (
	USDT   TokenSymbol = "USDT"
	USDC   TokenSymbol = "USDC"
	tether TokenSymbol = "Tether"
)

// Currencies supported by AkashicPay, includes native coins and tokens
const (
	CryptoUSDT CryptoCurrency = "USDT"
	CryptoUSDC CryptoCurrency = "USDC"
	CryptoTRX  CryptoCurrency = "TRX"
	CryptoETH  CryptoCurrency = "ETH"
	CryptoSEP  CryptoCurrency = "SEP"
)

type AkashicPay struct {
	TargetNode    acNode
	Env           Environment
	ApiSecret     string
	isFxBp        bool
	otk           Otk
	akashicUrl    string
	akashicPayUrl string
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
		return nil, newAkashicError(AkashicErrorCodeIsNotBp, "")
	}

	return &AkashicPay{
		TargetNode:    fastestNode,
		Env:           env,
		ApiSecret:     apiSecret,
		otk:           otk,
		isFxBp:        isBp.IsFxBp,
		akashicUrl:    urls.AkashicUrl,
		akashicPayUrl: urls.AkashicPayUrl,
	}, nil
}

// Get total balances, divided by Network and Token
func (ap *AkashicPay) GetBalance() ([]Balance, error) {
	ownerDetails, err := getBalance(ap.akashicUrl, ap.otk.Identity)

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

// Send a crypto-transaction
//
// referenceId is the userId or similar identifier for identifying the transaction
//
// to is the L1 or L2 address of the receiver
//
// Supply a zero-valued token to send native coin ("")
//
// The return is the L2 hash of the transaction
func (ap *AkashicPay) Payout(referenceId string, to string, amount string, network NetworkSymbol, token TokenSymbol) (string, error) {
	if referenceId == "" {
		return "", errors.New("referenceId may not be zero-valued")
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

	if err := validateDecimalPlaces(amount, network, token); err != nil {
		return "", err
	}

	ToAddress := to
	InitiatedToNonL2 := ""
	IsL2 := false

	DecimalAmount, err := convertToSmallestUnit(amount, network, token)
	if err != nil {
		return "", err
	}

	L2Lookup, err := ap.LookForL2Address(to, network)

	if err != nil {
		return "", err
	}

	InputIsL1, err := regexp.MatchString(networkDictionary[network].AddressRegex, to)
	if err != nil {
		return "", err
	}
	InputIsL2, err := regexp.MatchString(l2RegexWithOptionalPrefix, to)
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
			return "", newAkashicError(AkashicErrorCodeL2AddressNotFound, "")
		}
		IsL2 = true
	} else {
		// Must be alias
		if L2Lookup.L2Address == "" {
			return "", newAkashicError(AkashicErrorCodeL2AddressNotFound, "")
		}
		ToAddress = L2Lookup.L2Address
		InitiatedToNonL2 = to
		IsL2 = true
	}

	// L2
	if IsL2 {
		acToken := mapUSDTToTether(network, token)
		signedL2Tx, err := l2Transaction(ap.Env, ap.otk, network, DecimalAmount, ToAddress, acToken, InitiatedToNonL2, referenceId, ap.isFxBp)
		if err != nil {
			return "", err
		}

		//If FX, double-sign on BE
		if ap.isFxBp {
			res, err := prepareL2Txn(ap.akashicUrl, prepareL2TxnDto{SignedTx: signedL2Tx})
			if err != nil {
				return "", err
			}
			signedL2Tx = res.PreparedTxn
		}

		acRes, err := post[activeLedgerResponse[any, any]](ap.TargetNode.Node, signedL2Tx)
		if err != nil {
			return "", err
		}
		acErr := checkForAkashicChainError(acRes)
		if acErr != nil {
			return "", acErr
		}

		return prefixWithAS(acRes.Umid)
	}

	// L1

	Payload := prepareTxnDto{
		ToAddress:             to,
		Amount:                amount,
		NetworkSymbol:         network,
		TokenSymbol:           token,
		ReferenceId:           referenceId,
		Identity:              ap.otk.Identity,
		FeeDelegationStrategy: ffeeDelegationDelegate,
	}
	res, err := prepareL1Txn(ap.akashicUrl, Payload)

	if err != nil {
		if strings.Contains(err.Error(), "savingsExceeded") {
			return "", newAkashicError(AkashicErrorCodeSavingsExceeded, "")
		}
		return "", err
	}

	SignedTxn, err := signTransaction(res.PreparedTxn, ap.otk)

	if err != nil {
		return "", err
	}

	acRes, err := post[activeLedgerResponse[any, any]](ap.TargetNode.Node, SignedTxn)

	if err != nil {
		return "", err
	}
	acErr := checkForAkashicChainError(acRes)
	if acErr != nil {
		return "", acErr
	}

	return prefixWithAS(acRes.Umid)
}

// GetDepositUrl returns a url where a user can make deposits
//
// receiveCurrencies specifies which currencies you would like displayed as
// options on the page
//
// networks specifies which networks you would like displayed as options on the page
//
// referenceId is a parameter used to identify the order, can be left out ("")
//
// redirectUrl is a parameter which sets a URL to redirect to from the
// deposit URL, can be left out ("")
func (ap *AkashicPay) GetDepositUrl(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, networks []NetworkSymbol, redirectUrl string) (string, error) {
	return ap.getDepositUrlFunc(identifier, referenceId, receiveCurrencies, networks, redirectUrl, "", "", 0)
}

// Same as GetDepositUrl, but requires specifying the value of the deposit via
// requestedCurrency and requestedAmount
//
// unlike GetDepositUrl, referenceId must be specified
//
// Set the markupPercantage to adjust the exchange-rate for a markup/discount
func (ap *AkashicPay) GetDepositUrlWithRequestedValue(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, networks []NetworkSymbol, redirectUrl string, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (string, error) {
	if referenceId == "" {
		return "", errors.New("referenceId may not be zero-valued")
	}
	if requestedCurrency == "" {
		return "", errors.New("requestedCurrency may not be zero-valued")
	}
	if requestedAmount == "" {
		return "", errors.New("requestedAmount may not be zero-valued")
	}
	return ap.getDepositUrlFunc(identifier, referenceId, receiveCurrencies, networks, redirectUrl, requestedCurrency, requestedAmount, markupPercentage)
}

// GetDepositAddress returns an L1-address on the specified network for a user
// to deposit into
//
// referenceId is a parameter used to identify the order, can be left out ("")
func (ap *AkashicPay) GetDepositAddress(network NetworkSymbol, identifier string, referenceId string) (IDepositAddress, error) {
	return ap.getDepositAddressFunc(network, identifier, referenceId, "", "", "", 0)
}

// Same as GetDepositAddress, but requires specifying the value of the deposit via
// requestedCurrency and requestedAmount
//
// unlike GetDepositUrl, referenceId must be specified
//
// Set the markupPercantage to adjust the exchange-rate for a markup/discount
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

// GetExchangeRates return the exchange rates for all supported main-net coins
// in the value of the requested currency
func (ap *AkashicPay) GetExchangeRates(requestedCurrency Currency) (IGetExchangeRatesResult, error) {
	if requestedCurrency == "" {
		return IGetExchangeRatesResult{}, errors.New("requestedCurrency may not be zero-valued")
	}
	return getExchangeRates(ap.akashicUrl, requestedCurrency)
}

// LookForL2Address checks which L2-address an alias or L1-address belongs to.
// Or call with an L2-address to verify it exists
func (ap *AkashicPay) LookForL2Address(aliasOrL1OrL2Address string, network NetworkSymbol) (ILookForL2AddressResponse, error) {
	if aliasOrL1OrL2Address == "" {
		return ILookForL2AddressResponse{}, errors.New("aliasOrL1OrL2Address may not be zero-valued")
	}
	return getL2Lookup(ap.akashicUrl, aliasOrL1OrL2Address, network)
}

// Get all or a subset of transactions.
//
// Specify Page and Limit for pagination
func (ap *AkashicPay) GetTransfers(getTransactionParams IGetTransactions) ([]ITransaction, error) {
	validLimits := map[int]bool{0: true, 10: true, 25: true, 50: true, 100: true}
	if !validLimits[getTransactionParams.Limit] {
		return nil, errors.New("limit must be one of 10, 25, 50, or 100")
	}
	return getTransfers(ap.akashicUrl, ap.otk.Identity, getTransactionParams)
}

// GetTransactionDetails returns details about an individual transactions
//
// Returns an empty interface if no transaction found
func (ap *AkashicPay) GetTransactionDetails(l2Hash string) (ITransaction, error) {
	if l2Hash == "" {
		return ITransaction{}, errors.New("l2Hash may not be zero-valued")
	}
	return getTransactionDetails(ap.akashicUrl, l2Hash)
}

// Get the currently supported currencies in AkashicPay
// Returns a map from currencies to a list of networks supported for that currency
func (ap *AkashicPay) GetSupportedCurrencies() (map[CryptoCurrency][]NetworkSymbol, error) {
	return getSupportedCurrencies(ap.akashicUrl)
}

// VerifySignature can be used to verify a callback has not been altered. You
// must have initiated the SDK with your API-secret to do this
//
// Supply the callback-body as a string (`{"amount": "1", ...}`) and
// the signature from the callback-header
//
// Returns true if valid, indicating the callback has not been altered
func (ap *AkashicPay) VerifySignature(callback string, signature string) (bool, error) {
	if ap.ApiSecret == "" {
		return false, errors.New("apiSecret must be set if you want to verify a signature")
	}
	if !json.Valid([]byte(callback)) {
		return false, errors.New("callback is not valid JSON")
	}

	// Unmarshal and marshal again to sort the JSON alphabetically
	var i any
	err := json.Unmarshal([]byte(callback), &i)
	if err != nil {
		return false, err
	}

	sortedMsg, err := json.Marshal(i)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(ap.ApiSecret))
	_, err = mac.Write([]byte(sortedMsg))
	if err != nil {
		return false, err
	}

	expectedMAC := mac.Sum(nil)
	hexEncodedMAC := hex.EncodeToString(expectedMAC)
	return hexEncodedMAC == signature, nil
}

func chooseBestACNode(env Environment) (acNode, error) {
	var nodes map[string]acNode
	if env == Production {
		nodes = acNodes
	} else {
		nodes = acDevNodes
	}

	type Result struct {
		Node   acNode
		Status int
		Error  error
	}

	type AcResponse struct {
		Status int `json:"status"`
	}

	// Use a goroutine to concurrently check node health
	results := make(chan Result, len(nodes))
	for _, node := range nodes {
		go func(node acNode) {
			resp, err := http.Get(node.Node + "a/status")
			if err != nil {
				results <- Result{acNode{}, 0, err}
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				results <- Result{acNode{}, 0, err}
				return
			}

			var nodeData AcResponse
			err = json.Unmarshal(body, &nodeData)
			if err != nil {
				results <- Result{acNode{}, 0, err}
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
	return acNode{}, errors.New("no healthy AC node")
}

func (ap *AkashicPay) getDepositUrlFunc(identifier string, referenceId string, receiveCurrencies []CryptoCurrency, networks []NetworkSymbol, redirectUrl string, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (string, error) {
	if identifier == "" {
		return "", errors.New("identifier may not be zero-valued")
	}
	keys, err := getKeysByOwnerAndIdentifier(ap.akashicUrl, ap.otk.Identity, identifier)
	if err != nil {
		return "", err
	}
	preseedNetworks, err := ap.getPreseedNetworks()
	if err != nil {
		return "", err
	}

	// get networkSymbols that are owned
	existingSymbols := make(map[NetworkSymbol]bool)
	for _, key := range keys {
		existingSymbols[key.CoinSymbol] = true
	}

	// collect unassigned networks from supported networks
	var unassignedNetworks []NetworkSymbol
	for _, networkSymbol := range preseedNetworks {
		if _, exists := existingSymbols[networkSymbol]; !exists {
			unassignedNetworks = append(unassignedNetworks, networkSymbol)
			existingSymbols[networkSymbol] = true
		}
	}

	// bulk create or assign keys for unassigned networks
	if len(unassignedNetworks) > 0 {
		err := ap.bulkCreateOrAssignKeys(unassignedNetworks, identifier)
		if err != nil {
			return "", err
		}
	}

	if referenceId != "" {
		payload := iCreateDepositOrder{
			Identity:    ap.otk.Identity,
			ReferenceId: referenceId,
			Identifier:  identifier,
			Expires:     time.Now().UnixMilli() + 60*1000,
		}
		if markupPercentage != 0 {
			payload.MarkupPercentage = fmt.Sprintf("%f", markupPercentage)
		}
		if requestedAmount != "" && requestedCurrency != "" {
			payload.RequestedValue = &iRequestedValue{
				Amount:   requestedAmount,
				Currency: requestedCurrency,
			}
		}
		signature, err := signData(payload, ap.otk.privateKey)
		if err != nil {
			return "", err
		}
		payload.Signature = signature
		// create a deposit order
		_, err = createDepositOrder(ap.akashicUrl, payload)
		if err != nil {
			return "", err
		}
	}
	params := url.Values{}
	params.Set("identity", ap.otk.Identity)
	params.Set("identifier", identifier)
	if referenceId != "" {
		params.Set("referenceId", referenceId)
	}
	if len(receiveCurrencies) > 0 {
		params.Set("receiveCurrencies", strings.Join(cryptoCurrencySliceToStringSlice(receiveCurrencies), ","))
	}
	if len(networks) > 0 {
		params.Set("networks", strings.Join(networkSliceToStringSlice(networks), ","))
	}
	if redirectUrl != "" {
		params.Set("redirectUrl", base64.RawURLEncoding.EncodeToString([]byte(redirectUrl)))
	}
	return fmt.Sprintf("%v/sdk/deposit?%v", ap.akashicPayUrl, params.Encode()), nil
}

// createKey creates a new key on the specified network for the given identifier
// Returns the newly created key response
func (ap *AkashicPay) createKey(network NetworkSymbol, identifier string) (iKeyCreationResponse, error) {
	// Create a new key
	tx, err := keyCreateTransaction(ap.Env, network, ap.otk)
	if err != nil {
		return iKeyCreationResponse{}, err
	}

	createKeyRes, err := post[activeLedgerResponse[iKeyCreationResponse, any]](ap.TargetNode.Node, tx)
	if err != nil {
		return iKeyCreationResponse{}, err
	}
	acErr := checkForAkashicChainError(createKeyRes)
	if acErr != nil {
		return iKeyCreationResponse{}, acErr
	}
	newKey := createKeyRes.Responses[0]

	// Execute differential consensus transaction
	diffConTx, err := differentialConsensusTransaction(ap.Env, ap.otk, newKey, identifier)
	if err != nil {
		return iKeyCreationResponse{}, err
	}

	diffConTxResp, err := post[activeLedgerResponse[any, any]](ap.TargetNode.Node, diffConTx)
	if err != nil {
		return iKeyCreationResponse{}, err
	}
	acError := checkForAkashicChainError(diffConTxResp)
	if acError != nil {
		return iKeyCreationResponse{}, acError
	}

	return newKey, nil
}

// bulkCreateOrAssignKeys creates or assigns keys for multiple networks for a given identifier
// This function processes multiple networks and either creates new keys or assigns existing unassigned keys
func (ap *AkashicPay) bulkCreateOrAssignKeys(networks []NetworkSymbol, identifier string) error {
	var unassignedLedgerIds []string

	// Iterate through each network to check for existing keys or create new ones
	for _, network := range networks {
		response, err := getByOwnerAndIdentifier(ap.akashicUrl, network, identifier, ap.otk.Identity)
		if err != nil {
			return err
		}

		if response.UnassignedLedgerId != "" {
			// Collect unassigned ledger IDs for bulk assignment
			unassignedLedgerIds = append(unassignedLedgerIds, response.UnassignedLedgerId)
		} else if response.Address == "" && response.UnassignedLedgerId == "" {
			// If both do not exist, create new key and continue
			_, err := ap.createKey(network, identifier)
			if err != nil {
				return err
			}
		}
	}

	// If there are unassigned ledger IDs, assign them in bulk
	if len(unassignedLedgerIds) > 0 {
		tx, err := assign(ap.Env, ap.otk, unassignedLedgerIds, identifier)
		if err != nil {
			return err
		}

		// Assign keys to the user
		acRes, err := post[activeLedgerResponse[[]iKeyCreationResponse, any]](ap.TargetNode.Node, tx)
		if err != nil {
			return err
		}

		acErr := checkForAkashicChainError(acRes)
		if acErr != nil {
			return acErr
		}

		// Check if assignment was successful
		if len(acRes.Responses) == 0 {
			return newAkashicError(AkashicErrorCodeUnknownError, "Failed to assign keys for identifier "+identifier)
		}
	}

	return nil
}

func (ap *AkashicPay) getDepositAddressFunc(network NetworkSymbol, identifier string, referenceId string, token TokenSymbol, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (IDepositAddress, error) {
	// Check environment and network compatibility
	if (ap.Env == Development && (network == Ethereum_Mainnet || network == Tron)) ||
		(ap.Env == Production && (network == Ethereum_Sepolia || network == Tron_Shasta)) {
		return IDepositAddress{}, newAkashicError(AkashicErrorCodeNetworkEnvironmentMismatch, "")
	}
	if identifier == "" {
		return IDepositAddress{}, errors.New("identifier may not be zero-valued")
	}
	if network == "" {
		return IDepositAddress{}, errors.New("network may not be zero-valued")
	}
	response, err := getByOwnerAndIdentifier(ap.akashicUrl, network, identifier, ap.otk.Identity)
	if err != nil {
		return IDepositAddress{}, err
	}
	if response.Address != "" {
		if response.UnassignedLedgerId != "" {
			tx, err := assign(ap.Env, ap.otk, []string{response.UnassignedLedgerId}, identifier)
			if err != nil {
				return IDepositAddress{}, err
			}

			acRes, err := post[activeLedgerResponse[iKeyCreationResponse, any]](ap.TargetNode.Node, tx)
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
	newKey, err := ap.createKey(network, identifier)
	if err != nil {
		return IDepositAddress{}, err
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

func (ap *AkashicPay) createDepositPayloadAndOrder(referenceId string, identifier string, address string, network NetworkSymbol, tokenSymbol TokenSymbol, requestedCurrency Currency, requestedAmount string, markupPercentage float64) (iCreateDepositOrderResponse, error) {
	payload := iCreateDepositOrder{
		Identity:    ap.otk.Identity,
		Expires:     time.Now().UnixMilli() + 60*1000,
		ReferenceId: referenceId,
		Identifier:  identifier,
		ToAddress:   address,
		CoinSymbol:  network,
		TokenSymbol: tokenSymbol,
	}

	if markupPercentage != 0 {
		payload.MarkupPercentage = fmt.Sprintf("%f", markupPercentage)
	}

	if requestedCurrency != "" && requestedAmount != "" {
		payload.RequestedValue = &iRequestedValue{
			Currency: requestedCurrency,
			Amount:   requestedAmount,
		}
	}

	signature, err := signData(payload, ap.otk.privateKey)
	if err != nil {
		return iCreateDepositOrderResponse{}, err
	}
	createOrderPayload := payload
	createOrderPayload.Signature = signature
	return createDepositOrder(ap.akashicUrl, createOrderPayload)
}

// getPreseedNetworks returns a list of networks that need to create key or assign preseed keys
func (ap *AkashicPay) getPreseedNetworks() ([]NetworkSymbol, error) {
	supportedCurrencies, err := getSupportedCurrencies(ap.akashicUrl)
	if err != nil {
		return nil, err
	}

	duplicatedSymbols := make(map[NetworkSymbol]bool)
	var preseedNetworks []NetworkSymbol
	// flatten all supported networks and deduplicate
	for _, networkSymbols := range supportedCurrencies {
		for _, networkSymbol := range networkSymbols {
			if _, duplicated := duplicatedSymbols[networkSymbol]; !duplicated {
				// Only add to preseedNetworks if networkSymbol is NOT in NonEthEvmNetworks
				isNonEthEvm := false
				for _, nonEthEvmNetwork := range NonEthEvmNetworks {
					if networkSymbol == nonEthEvmNetwork {
						isNonEthEvm = true
						break
					}
				}
				if !isNonEthEvm {
					preseedNetworks = append(preseedNetworks, networkSymbol)
				}
				duplicatedSymbols[networkSymbol] = true
			}
		}
	}
	return preseedNetworks, nil
}
