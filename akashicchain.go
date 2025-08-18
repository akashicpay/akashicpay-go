package akashicpay

import (
	"encoding/hex"
	"encoding/json"
	"slices"
	"strings"
	"time"

	alsdk "github.com/activeledger/SDK-Golang"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/titanous/bitcoin-crypto/bitecdsa"
	"github.com/titanous/bitcoin-crypto/bitelliptic"
)

const acNamespace = "akashicchain"
const acNativeCoin = "#native"

// contracts
type contracts struct {
	FxMultiSigner      string // Not a contract
	Namespace          string // Also not a contract
	Create             string
	CryptoTransfer     string
	DiffConsensus      string
	Onboard            string
	AssignKey          string
	CreateSecondaryOtk string
}

var acMainNetContracts = contracts{
	FxMultiSigner:  "ASad1414566948845b404e8b6ac91639cc3643129d0ef8b7828ede7a0ac1044d6e",
	Create:         "50e1372f0d3805dac4a51299bb0e99960862d7d01f247e85725d99011682b8ac@1",
	CryptoTransfer: "2bae6ea681826c0307ee047ef68eb0cf53487a257c498de7d081d66de119d666@1",
	DiffConsensus:  "94479927cbe0860a3f51cbd36230faef7d1b69974323a83c8abcc78e3d0e8dd9@1",
	Onboard:        "a456ddc07da6d46a6897d24de188e767b87a9d9f2f3c617d858aaf819e0e5bce@1",
	AssignKey:      "7afea15e8028af9f5aeafb6db6c0d2e8969c0c0492360ab15a6bc3754b818e19@1",
}

var acTestNetContracts = contracts{
	FxMultiSigner:  "ASeffcb8790aff2439522ef4bd834cca5233dc1670e5fa1c93fa19305323937a17",
	Create:         "ad171259a7c628ba6993c6bd555f07111525128194aa4226662e48a0b0a93116@1",
	CryptoTransfer: "a32a8bc21ceaeeaa671573126a246c15ec4dc3a5c825e3cffc9441636019acb1@1",
	DiffConsensus:  "17be1db84dbf81c1ff1b2f5aebd4ba4e95d81338daf98d7c2bc7b54ad8994d1c@1",
	Onboard:        "c19c6f4d3c443ae7abb14d17d33b29d134df8d11bdabc568bd23f7023ee991fd@1",
	AssignKey:      "a6e95e2f563bdac69bfa265b1c215bf2125e1c50048f68f9c0b52982e320d675@1",
}

type acTransaction struct {
	TxObject  txObject               `json:"$tx"`
	SelfSign  bool                   `json:"$selfsign"`
	Signature map[string]interface{} `json:"$sigs"`
	Unanimous bool                   `json:"$unanimous"`
}

// txObject within a transaction
type txObject struct {
	Namespace string                 `json:"$namespace"`
	Contract  string                 `json:"$contract"`
	Entry     string                 `json:"$entry,omitempty"`
	Input     map[string]interface{} `json:"$i"`
	Output    map[string]interface{} `json:"$o,omitempty"`
	ReadOnly  map[string]interface{} `json:"$r,omitempty"`
	DbIndex   int                    `json:"_dbIndex,omitempty"`
	Expire    string                 `json:"$expire,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type iKeyCreationResponse struct {
	Id      string   `json:"id"`
	Address string   `json:"address"`
	Hashes  []string `json:"hashes"`
}

// AC Response
type activeLedgerResponse[ResponseT, StreamT any] struct {
	Umid    string `json:"$umid"`
	Summary struct {
		Total  int      `json:"total"`
		Vote   int      `json:"vote"`
		Commit int      `json:"commit"`
		Errors []string `json:"errors,omitempty"`
	} `json:"$summary"`
	Streams struct {
		New     []StreamT `json:"new"`
		Updated []StreamT `json:"updated"`
	} `json:"$streams"`
	Responses []ResponseT `json:"$responses,omitempty"`
	Debug     any         `json:"$debug,omitempty"`
}

// getACSymbol returns the canonical symbol for a given network symbol (e.g., 'trx' for Tron, 'eth' for Ethereum)
func getACSymbol(coinSymbol NetworkSymbol) string {
	switch coinSymbol {
	case Tron, Tron_Shasta:
		return "trx"
	case Ethereum_Mainnet, Ethereum_Sepolia:
		return "eth"
	default:
		return string(coinSymbol)
	}
}

// getACNetwork returns the canonical network name for a given network symbol
func getACNetwork(coinSymbol NetworkSymbol) string {
	switch coinSymbol {
	case Ethereum_Mainnet:
		return "ETH"
	case Ethereum_Sepolia:
		return "SEP"
	case Tron:
		return "trx"
	case Tron_Shasta:
		return "shasta"
	default:
		return string(coinSymbol)
	}
}

// signData signs any data structure using the provided private key and returns the signature string.
func signData(data interface{}, privateKey string) (string, error) {
	txObjectByte, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	privateKeyByte, err := hex.DecodeString(strings.TrimLeft(privateKey, "0x"))
	if err != nil {
		return "", err
	}
	priv, pub := btcec.PrivKeyFromBytes([]byte(privateKeyByte))
	privECDSA := priv.ToECDSA()
	pubECDSA := pub.ToECDSA()
	privateKeyObj := new(bitecdsa.PrivateKey)
	privateKeyObj.PublicKey.BitCurve = bitelliptic.S256()
	privateKeyObj.D = privECDSA.D
	privateKeyObj.PublicKey.X = pubECDSA.X
	privateKeyObj.PublicKey.Y = pubECDSA.Y
	sign := alsdk.EcdsaSign(privateKeyObj, string(txObjectByte))
	return sign, nil
}

// signTransaction signs an ACTransaction and attaches the signature to its Signature map.
func signTransaction(tx acTransaction, otk Otk) (acTransaction, error) {
	sign, err := signData(tx.TxObject, otk.privateKey)
	if err != nil {
		return acTransaction{}, err
	}
	tx.Signature[otk.Identity] = sign
	return tx, nil
}

func keyCreateTransaction(env Environment, coinSymbol NetworkSymbol, otk Otk) (acTransaction, error) {
	contracts := acTestNetContracts
	dbIndex := 15

	if env == Production {
		contracts = acMainNetContracts
		dbIndex = 0
	}
	TxBody := acTransaction{
		TxObject: txObject{
			Namespace: acNamespace,
			Contract:  contracts.Create,
			Input: map[string]interface{}{
				"owner": map[string]interface{}{
					"$stream":  otk.Identity,
					"symbol":   getACSymbol(coinSymbol),
					"network":  getACNetwork(coinSymbol),
					"business": true,
				},
			},
			DbIndex: dbIndex,
		},
		Signature: map[string]interface{}{},
	}
	addExpireToTx(&TxBody)
	return signTransaction(TxBody, otk)
}

// Create and Sign an L2 transaction
func l2Transaction(
	env Environment,
	otk Otk,
	coinSymbol NetworkSymbol,
	amount string,
	toAddress string,
	tokenSymbol TokenSymbol,
	initiatedToNonL2 string,
	referenceId string,
	isFxBp bool,
) (acTransaction, error) {
	DbIndex := 15
	Contracts := acTestNetContracts
	Token := acNativeCoin
	Metadata := map[string]interface{}{
		"referenceId": referenceId,
	}

	if env == Production {
		DbIndex = 0
		Contracts = acMainNetContracts
	}

	if tokenSymbol != "" {
		Token = string(tokenSymbol)
	}

	Input := map[string]interface{}{
		"owner": map[string]interface{}{
			"$stream": otk.Identity,
			"network": coinSymbol,
			"token":   Token,
			"amount":  amount,
		},
	}
	if isFxBp {
		Input["afx"] = map[string]interface{}{
			"$stream":  Contracts.FxMultiSigner,
			"$sigOnly": true,
		}
	}
	if initiatedToNonL2 != "" {
		Metadata["initiatedToNonL2"] = initiatedToNonL2
	}

	TxBody := acTransaction{
		TxObject: txObject{
			Namespace: acNamespace,
			Contract:  Contracts.CryptoTransfer,
			Entry:     "transfer",
			Input:     Input,
			Output: map[string]interface{}{
				"to": map[string]interface{}{
					"$stream": toAddress,
				},
			},
			DbIndex:  DbIndex,
			Metadata: Metadata,
		},
		Signature: map[string]interface{}{},
	}
	addExpireToTx(&TxBody)
	return signTransaction(TxBody, otk)
}

func addExpireToTx(tx *acTransaction) *acTransaction {
	if tx.TxObject.Expire != "" {
		return tx
	}
	tx.TxObject.Expire = time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	return tx
}

func checkForAkashicChainError[T any](response activeLedgerResponse[T, any]) error {
	if response.Summary.Commit > 0 {
		return nil
	}
	acErrorString := response.Summary.Errors[0]

	if slices.ContainsFunc(
		[]string{"balance is not sufficient", "Couldn't parse integer", "Part-Balance to low"},
		func(e string) bool { return strings.Contains(acErrorString, e) },
	) {
		return newAkashicError(AkashicErrorCodeSavingsExceeded, "")
	}
	if strings.Contains(acErrorString, "Stream(s) not found") {
		return newAkashicError(AkashicErrorCodeL2AddressNotFound, "")
	}
	return newAkashicError(AkashicErrorCodeUnknownError, "AkashicChain Failure: "+acErrorString)
}

// AssignKeyTransaction creates and signs a transaction to assign a key to a user identifier
func assign(env Environment, otk Otk, ledgerId string, identifier string) (acTransaction, error) {
	contracts := acTestNetContracts
	dbIndex := 15

	if env == Production {
		contracts = acMainNetContracts
		dbIndex = 0
	}

	TxBody := acTransaction{
		TxObject: txObject{
			Namespace: acNamespace,
			Contract:  contracts.AssignKey,
			Input: map[string]interface{}{
				"owner": map[string]interface{}{
					"$stream": otk.Identity,
				},
			},
			Output: map[string]interface{}{
				"key": map[string]interface{}{
					"$stream": ledgerId,
				},
			},
			Metadata: map[string]interface{}{
				"identifier": identifier,
			},
			DbIndex: dbIndex,
		},
		Signature: map[string]interface{}{},
	}
	addExpireToTx(&TxBody)
	return signTransaction(TxBody, otk)
}

func differentialConsensusTransaction(
	env Environment,
	otk Otk,
	key iKeyCreationResponse,
	identifier string,
) (acTransaction, error) {
	contracts := acTestNetContracts
	dbIndex := 15

	if env == Production {
		contracts = acMainNetContracts
		dbIndex = 0
	}

	TxBody := acTransaction{
		TxObject: txObject{
			Namespace: acNamespace,
			Contract:  contracts.DiffConsensus,
			Input: map[string]interface{}{
				"owner": map[string]interface{}{
					"publicKey": otk.publicKey,
					"type":      "secp256k1",
					"address":   key.Address,
					"hashes":    key.Hashes,
				},
			},
			Output: map[string]interface{}{
				"key": map[string]interface{}{
					"$stream": key.Id,
				},
			},
			DbIndex: dbIndex,
			Metadata: map[string]interface{}{
				"identifier": identifier,
			},
		},
		SelfSign:  true,
		Signature: map[string]interface{}{},
		Unanimous: true,
	}

	addExpireToTx(&TxBody)
	newOtk := otk
	newOtk.Identity = "owner"
	return signTransaction(TxBody, newOtk)
}
