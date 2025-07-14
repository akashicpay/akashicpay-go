package akashicpay

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	alsdk "github.com/activeledger/SDK-Golang"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/titanous/bitcoin-crypto/bitecdsa"
	"github.com/titanous/bitcoin-crypto/bitelliptic"
)

type Transaction struct {
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

// TODO: This is ugly, but proves the concept (I was able to make a key!)
// Should factor out all the signing bits into a separate func
func KeyCreateTransaction(coinSymbol NetworkSymbol, otk Otk) (Transaction, error) {
	TxBody := Transaction{
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
	txObjectByte, _ := json.Marshal(TxBody.TxObject)

	privateKeyByte, err := hex.DecodeString(strings.TrimLeft(otk.privateKey, "0x"))

	if err != nil {
		return Transaction{}, err
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
	TxBody.Signature = map[string]interface{}{
		otk.Identity: sign,
	}
	return TxBody, nil
}
