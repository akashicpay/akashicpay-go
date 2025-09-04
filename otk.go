package akashicpay

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
)

// private key could be 62 or 64 characters long despite being prefixed with 0x
const acPrivateKeyRegex = `^0x[a-f\d]{62,64}$`

type Otk struct {
	privateKey string
	publicKey  string
	Identity   string
}

func reconstructOtkFromPrivateKey(privateKey string, identity string) (Otk, error) {
	matchesRegex, _ := regexp.MatchString(acPrivateKeyRegex, privateKey)
	if !matchesRegex {
		return Otk{}, newAkashicError(AkashicErrorCodeIncorrectPrivateKeyFormat, "")
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
