package queries

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/eth_client"
	"koonopek/know_your_rpc/server/server"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const USDC = 1e6

type PaymentConfig struct {
	ExpectedValue *big.Int       `json:"expectedValue"`
	ExpectedToken common.Address `json:"expectedToken"`
	ExpectedTo    common.Address `json:"expectedTo"`
	ChainID       int64          `json:"chainId"`
}

var paymentConfig = PaymentConfig{
	ExpectedValue: big.NewInt(10 * USDC), // 10 USDC
	ExpectedToken: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
	ExpectedTo:    common.HexToAddress("0x69dF8F2010843dA5Bfe3df08aB769940764Bb64f"),
	ChainID:       1,
}

type AcknowledgePaymentRequest struct {
	TxHash string `json:"txHash"`
}

func CreateGetPaymentDataQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		WriteHttpResponse(paymentConfig, w)
	}
}

func CreateGetSubscriptionQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		requestAuthorizerAddress, isNotAuthorized := GetRequestAuthorizerAddress(r, w)

		if isNotAuthorized {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		server.LockUserStorageMutex(requestAuthorizerAddress)
		defer server.UnlockUserStorageMutex(requestAuthorizerAddress)
		userStore, err := server.ReadAndUpdateUserData(requestAuthorizerAddress)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		WriteHttpResponse(userStore.Subscription, w)
	}
}

func CreateAcknowledgePaymentQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var request AcknowledgePaymentRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		requestAuthorizerAddress, isNotAuthorized := GetRequestAuthorizerAddress(r, w)

		if requestAuthorizerAddress == "public" {
			log.Printf("failed to acknowledge payment because Authorization header is not set")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if isNotAuthorized {
			log.Printf("failed to acknowledge payment because you are not authorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = eth_client.VerifyErc20Transfer(serverContext.EthClient, request.TxHash, common.HexToAddress(requestAuthorizerAddress), paymentConfig.ExpectedTo, paymentConfig.ExpectedValue, paymentConfig.ExpectedToken)
		if err != nil {
			log.Printf("failed to acknowledge payment because %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		server.LockUserStorageMutex(requestAuthorizerAddress)
		defer server.UnlockUserStorageMutex(requestAuthorizerAddress)
		server.InvalidateUserDataCache(requestAuthorizerAddress)
		userStore, err := server.ReadAndUpdateUserData(requestAuthorizerAddress)
		if err != nil {
			log.Printf("failed to acknowledge payment because %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userStore.Subscription.TxReceipts = append(userStore.Subscription.TxReceipts, types.TxReceipt{
			TxHash:  request.TxHash,
			ChainID: paymentConfig.ChainID,
		})

		//TODO: if already validUntil is set, then increase it by 30 days
		userStore.Subscription.ValidUntil = time.Now().Add(time.Hour * 24 * 30).Unix()

		updatedData, err := json.Marshal(userStore)
		if err != nil {
			log.Printf("failed to acknowledge payment because %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = s3.WriteS3Object(server.USERS_BUCKET, fmt.Sprintf("%s.json", requestAuthorizerAddress), updatedData)
		if err != nil {
			log.Printf("failed to acknowledge payment because %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
