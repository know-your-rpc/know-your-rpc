package queries

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func ExtractSigner(sigHex string, msgStr string) (string, error) {
	sig := hexutil.MustDecode(sigHex)
	msg := []byte(msgStr)

	msg = accounts.TextHash(msg)
	if sig[crypto.RecoveryIDOffset] == 27 || sig[crypto.RecoveryIDOffset] == 28 {
		sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(msg, sig)
	if err != nil {
		return "", fmt.Errorf("failed to recover")
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)

	return recoveredAddr.Hex(), nil
}
