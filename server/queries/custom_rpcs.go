package queries

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
	"net/http"
)

func handleCustomRpcRequest(w http.ResponseWriter, r *http.Request, operation func(*types.RpcInfoMap, string, string) error) {
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

	userAddress, isNotAuthorized := GetAuthorizedBucketKey(r, w)
	if isNotAuthorized || userAddress == server.PUBLIC_S3_KEY {
		fmt.Printf("unauthorized request bucket_key=%s \n", userAddress)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketKey := fmt.Sprintf("%s.json", userAddress)

	data, err := s3.ReadS3Object(server.USERS_BUCKET, bucketKey)
	if err != nil {
		fmt.Printf("failed to read from s3 bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rpcInfoMap := &types.RpcInfoMap{}
	if err := json.Unmarshal(data, rpcInfoMap); err != nil {
		fmt.Printf("failed to unmarshal rpc info map bucket_key=%s \n", bucketKey)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := operation(rpcInfoMap, request.ChainId, request.RpcUrl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updatedData, err := json.Marshal(rpcInfoMap)
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

	w.WriteHeader(http.StatusOK)
}

func CreateCustomRpcAddQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(rpcInfoMap *types.RpcInfoMap, chainId, rpcUrl string) error {
			chainRpcs, exists := (*rpcInfoMap)[chainId]
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
			(*rpcInfoMap)[chainId] = append(chainRpcs, newRpc)
			return nil
		})
	}
}

func CreateCustomRpcRemoveQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(rpcInfoMap *types.RpcInfoMap, chainId, rpcUrl string) error {
			chainRpcs, exists := (*rpcInfoMap)[chainId]
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

			(*rpcInfoMap)[chainId] = newChainRpcs
			return nil
		})
	}
}

func CreateCustomRpcRemoveAllQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handleCustomRpcRequest(w, r, func(rpcInfoMap *types.RpcInfoMap, chainId, _ string) error {
			_, exists := (*rpcInfoMap)[chainId]
			if !exists {
				fmt.Printf("chain id not found chain_id=%s \n", chainId)
				return fmt.Errorf("chain id not found")
			}

			// Remove all RPCs for the specified chain
			(*rpcInfoMap)[chainId] = []types.RpcInfo{}
			return nil
		})
	}
}
