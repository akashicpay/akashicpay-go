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

const ACNamespace = "akashicchain"
const ACNativeCoin = "#native"

// Contracts
type Contracts struct {
	FxMultiSigner      string // Not a contract
	Namespace          string // Also not a contract
	Create             string
	CryptoTransfer     string
	DiffConsensus      string
	Onboard            string
	AssignKey          string
	CreateSecondaryOtk string
}

var ACMainNetContracts = Contracts{
	FxMultiSigner:  "ASad1414566948845b404e8b6ac91639cc3643129d0ef8b7828ede7a0ac1044d6e",
	Create:         "50e1372f0d3805dac4a51299bb0e99960862d7d01f247e85725d99011682b8ac@1",
	CryptoTransfer: "2bae6ea681826c0307ee047ef68eb0cf53487a257c498de7d081d66de119d666@1",
	DiffConsensus:  "94479927cbe0860a3f51cbd36230faef7d1b69974323a83c8abcc78e3d0e8dd9@1",
	Onboard:        "a456ddc07da6d46a6897d24de188e767b87a9d9f2f3c617d858aaf819e0e5bce@1",
	AssignKey:      "7afea15e8028af9f5aeafb6db6c0d2e8969c0c0492360ab15a6bc3754b818e19@1",
}

var ACTestNetContracts = Contracts{
	FxMultiSigner:  "ASeffcb8790aff2439522ef4bd834cca5233dc1670e5fa1c93fa19305323937a17",
	Create:         "ad171259a7c628ba6993c6bd555f07111525128194aa4226662e48a0b0a93116@1",
	CryptoTransfer: "a32a8bc21ceaeeaa671573126a246c15ec4dc3a5c825e3cffc9441636019acb1@1",
	DiffConsensus:  "17be1db84dbf81c1ff1b2f5aebd4ba4e95d81338daf98d7c2bc7b54ad8994d1c@1",
	Onboard:        "c19c6f4d3c443ae7abb14d17d33b29d134df8d11bdabc568bd23f7023ee991fd@1",
	AssignKey:      "a6e95e2f563bdac69bfa265b1c215bf2125e1c50048f68f9c0b52982e320d675@1",
}

type ACTransaction struct {
	TxObject  TxObject               `json:"$tx"`
	SelfSign  bool                   `json:"$selfsign"`
	Signature map[string]interface{} `json:"$sigs"`
}

// TxObject within a transaction
type TxObject struct {
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

// AC Response
type ActiveLedgerResponse[ResponseT, StreamT any] struct {
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

// TODO: This is ugly, but proves the concept (I was able to make a key!)
// Should factor out all the signing bits into a separate func
func KeyCreateTransaction(coinSymbol NetworkSymbol, otk Otk) (ACTransaction, error) {
	TxBody := ACTransaction{
		TxObject: TxObject{
			Namespace: "akashicchain",
			Contract:  "ad171259a7c628ba6993c6bd555f07111525128194aa4226662e48a0b0a93116@1",
			Input: map[string]interface{}{
				"owner": map[string]interface{}{
					"$stream":  otk.Identity,
					"symbol":   "trx",
					"network":  "shasta",
					"business": true,
				},
			},
			DbIndex: 15,
		},
	}
	return SignTransaction(TxBody, otk)
}

// Create and Sign an L2 transaction
func L2Transaction(
	env Environment,
	otk Otk,
	coinSymbol NetworkSymbol,
	amount string,
	toAddress string,
	tokenSymbol TokenSymbol,
	initiatedToNonL2 string,
	identifier string,
	isFxBp bool,
) (ACTransaction, error) {
	DbIndex := 15
	Contracts := ACTestNetContracts
	Token := ACNativeCoin
	Metadata := map[string]interface{}{
		"identifier": identifier,
	}

	if env == Production {
		DbIndex = 0
		Contracts = ACMainNetContracts
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

	TxBody := ACTransaction{
		TxObject: TxObject{
			Namespace: ACNamespace,
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
	return SignTransaction(TxBody, otk)
}

func addExpireToTx(tx *ACTransaction) *ACTransaction {
	if tx.TxObject.Expire != "" {
		return tx
	}
	tx.TxObject.Expire = time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	return tx
}

func SignTransaction(tx ACTransaction, otk Otk) (ACTransaction, error) {
	txObjectByte, _ := json.Marshal(tx.TxObject)

	privateKeyByte, err := hex.DecodeString(strings.TrimLeft(otk.privateKey, "0x"))

	if err != nil {
		return ACTransaction{}, err
	}
	priv, pub := btcec.PrivKeyFromBytes([]byte(privateKeyByte))

	privECDSA := priv.ToECDSA()
	pubECDSA := pub.ToECDSA()

	privateKey := new(bitecdsa.PrivateKey)
	privateKey.PublicKey.BitCurve = bitelliptic.S256()
	privateKey.D = privECDSA.D
	privateKey.PublicKey.X = pubECDSA.X
	privateKey.PublicKey.Y = pubECDSA.Y

	sign := alsdk.EcdsaSign(privateKey, string(txObjectByte))
	tx.Signature[otk.Identity] = sign
	return tx, nil
}

func checkForAkashicChainError(response ActiveLedgerResponse[any, any]) error {
	if response.Summary.Commit > 0 {
		return nil
	}
	acErrorString := response.Summary.Errors[0]

	if slices.ContainsFunc(
		[]string{"balance is not sufficient", "Couldn't parse integer", "Part-Balance to low"},
		func(e string) bool { return strings.Contains(acErrorString, e) },
	) {
		return NewAkashicError(AkashicErrorCodeSavingsExceeded, "")
	}
	if strings.Contains(acErrorString, "Stream(s) not found") {
		return NewAkashicError(AkashicErrorCodeL2AddressNotFound, "")
	}
	return NewAkashicError(AkashicErrorCodeUnknownError, "AkashicChain Failure: "+acErrorString)
}
