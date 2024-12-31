package queries

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
	"net/http"
)

func handleCustomRpcRequest(w http.ResponseWriter, r *http.Request, operation func(*types.UserData, string, string) error) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request struct {
		RpcUrl  string `json:"rpcUrl" validate:"required,url"`
		ChainId string `json:"chainId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Printf("failed to decode request err=%s \n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := Validator.Struct(request); err != nil {
		fmt.Printf("invalid input err=%s \n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationHeader := r.Header.Get("Authorization")

	userAddress, err := ExtractSignerFromAuthHeader(authorizationHeader)
	if err != nil {
		fmt.Printf("unauthorized request err=%s \n", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bucketKey := fmt.Sprintf("%s.json", userAddress)

	server.LockUserStorageMutex(userAddress)
	defer server.UnlockUserStorageMutex(userAddress)
	//TODO: why do we read here straight from the bucket?
	data, err := s3.ReadS3Object(server.USERS_BUCKET, bucketKey)
	if err != nil {
		fmt.Printf("failed to read from s3 bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userStore := &types.UserData{}
	if err := json.Unmarshal(data, userStore); err != nil {
		fmt.Printf("failed to unmarshal rpc info map bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !userStore.IsSubscriptionValid() {
		fmt.Printf("subscription is not valid user_address=%s expired_at=%d \n", userAddress, userStore.Subscription.ValidUntil)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := operation(userStore, request.ChainId, request.RpcUrl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedData, err := json.Marshal(userStore)
	if err != nil {
		fmt.Printf("failed to marshal rpc info map bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s3.WriteS3Object(server.USERS_BUCKET, bucketKey, updatedData); err != nil {
		fmt.Printf("failed to write to s3 bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	server.InvalidateUserDataCache(userAddress)

	w.WriteHeader(http.StatusOK)
}

func CreateCustomRpcAddQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(userStore *types.UserData, chainId, rpcUrl string) error {
			chainRpcs, exists := userStore.RpcInfo[chainId]
			if !exists {
				chainRpcs = []types.RpcInfo{}
			}

			for _, rpc := range chainRpcs {
				if rpc.URL == rpcUrl {
					fmt.Printf("rpc url already exists chain_id=%s rpc_url=%s \n", chainId, rpcUrl)
					return fmt.Errorf("rpc url already exists")
				}
			}

			newRpc := types.RpcInfo{URL: rpcUrl}
			userStore.RpcInfo[chainId] = append(chainRpcs, newRpc)
			return nil
		})
	}
}

func CreateCustomRpcRemoveQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(userStore *types.UserData, chainId, rpcUrl string) error {
			chainRpcs, exists := userStore.RpcInfo[chainId]
			if !exists {
				fmt.Printf("chain id not found chain_id=%s \n", chainId)
				return fmt.Errorf("chain id not found")
			}

			found := false
			newChainRpcs := []types.RpcInfo{}
			for _, rpc := range chainRpcs {
				if rpc.URL != rpcUrl {
					newChainRpcs = append(newChainRpcs, rpc)
				} else {
					found = true
				}
			}

			if !found {
				fmt.Printf("rpc url not found chain_id=%s rpc_url=%s \n", chainId, rpcUrl)
				return fmt.Errorf("rpc url not found")
			}

			userStore.RpcInfo[chainId] = newChainRpcs
			return nil
		})
	}
}

func CreateCustomRpcRemoveAllQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(userStore *types.UserData, chainId, _ string) error {
			_, exists := userStore.RpcInfo[chainId]
			if !exists {
				fmt.Printf("chain id not found chain_id=%s \n", chainId)
				return fmt.Errorf("chain id not found")
			}

			// Remove all RPCs for the specified chain
			userStore.RpcInfo[chainId] = []types.RpcInfo{}
			return nil
		})
	}
}

func CreateCustomRpcSyncQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var request struct {
			ChainId string   `json:"chainId" validate:"required"`
			RpcUrls []string `json:"rpcUrls" validate:"required,dive,url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			fmt.Printf("failed to decode request err=%s \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := Validator.Struct(request); err != nil {
			fmt.Printf("invalid input err=%s \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		authorizationHeader := r.Header.Get("Authorization")
		userAddress, err := ExtractSignerFromAuthHeader(authorizationHeader)
		if err != nil {
			fmt.Printf("unauthorized request err=%s \n", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		bucketKey := fmt.Sprintf("%s.json", userAddress)

		server.LockUserStorageMutex(userAddress)
		defer server.UnlockUserStorageMutex(userAddress)

		data, err := s3.ReadS3Object(server.USERS_BUCKET, bucketKey)
		if err != nil {
			fmt.Printf("failed to read from s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userStore := &types.UserData{}
		if err := json.Unmarshal(data, userStore); err != nil {
			fmt.Printf("failed to unmarshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !userStore.IsSubscriptionValid() {
			fmt.Printf("subscription is not valid user_address=%s expired_at=%d \n", userAddress, userStore.Subscription.ValidUntil)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Convert string URLs to RpcInfo structs
		newRpcs := make([]types.RpcInfo, len(request.RpcUrls))
		for i, url := range request.RpcUrls {
			newRpcs[i] = types.RpcInfo{URL: url}
		}

		// Replace existing RPCs with new ones
		userStore.RpcInfo[request.ChainId] = newRpcs

		updatedData, err := json.Marshal(userStore)
		if err != nil {
			fmt.Printf("failed to marshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := s3.WriteS3Object(server.USERS_BUCKET, bucketKey, updatedData); err != nil {
			fmt.Printf("failed to write to s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		server.InvalidateUserDataCache(userAddress)
		w.WriteHeader(http.StatusOK)
	}
}
