package queries

import (
	"fmt"
	"net/http"
	"strings"

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

func GetAuthorizedBucketKey(r *http.Request, w http.ResponseWriter) (string, bool) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader != "" && authorizationHeader != "undefined" {
		splitted := strings.Split(authorizationHeader, "#")
		if len(splitted) != 2 {
			fmt.Printf("wrong formatted authorization header authorization_header=%s \n", authorizationHeader)
			w.WriteHeader(http.StatusBadRequest)
			return "", true
		}
		signature := splitted[0]
		msg := splitted[1]
		signer, err := ExtractSigner(signature, msg)
		if err != nil {
			fmt.Printf("failed to extract signer")
			w.WriteHeader(http.StatusUnauthorized)
			return "", true
		}

		return signer, false
	}
	return "public", false
}
