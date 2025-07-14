package akashicpay

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
)

const ACPrivateKeyRegex = `^0x[a-f\d]{64}$`

type Otk struct {
	privateKey string
	publicKey  string
	Identity   string
}

func reconstructOtkFromPrivateKey(privateKey string, identity string) (Otk, error) {
	matchesRegex, _ := regexp.MatchString(ACPrivateKeyRegex, privateKey)
	if !matchesRegex {
		return Otk{}, NewAkashicError(AkashicErrorCodeIncorrectPrivateKeyFormat, "")
	}
	// Remove 0x if there
	privateKey = strings.TrimLeft(privateKey, "0x")

	privateKeyByte, err := hex.DecodeString(privateKey)

	if err != nil {
		return Otk{}, fmt.Errorf("failed to decode private key: %w", err)
	}
	// Gets Public key, puts it in hex and adds the 0x
	_, pub := btcec.PrivKeyFromBytes([]byte(privateKeyByte))

	publicKey := "0x" + hex.EncodeToString(pub.SerializeCompressed())

	return Otk{
		privateKey: privateKey,
		publicKey:  publicKey,
		Identity:   identity,
	}, nil
}
