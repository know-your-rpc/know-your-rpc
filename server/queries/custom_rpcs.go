package queries

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
	"net/http"
)

type CustomRpcAddRequest struct {
	RpcUrl  string `json:"rpcUrl" validate:"required,url"`
	ChainId string `json:"chainId" validate:"required"`
}

func CreateCustomRpcAddQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		request := CustomRpcAddRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			fmt.Printf("failed to decode request err=%s \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate the struct fields
		err = Validator.Struct(request)
		if err != nil {
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

		// Read the existing RPC info from S3
		data, err := s3.ReadS3Object(server.USERS_BUCKET, bucketKey)
		if err != nil {
			fmt.Printf("failed to read from s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rpcInfoMap := &types.RpcInfoMap{}
		err = json.Unmarshal(data, rpcInfoMap)
		if err != nil {
			fmt.Printf("failed to unmarshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if the chain ID exists in the map
		chainRpcs, exists := (*rpcInfoMap)[request.ChainId]
		if !exists {
			chainRpcs = []types.RpcInfo{}
		}

		// Check if the RPC URL already exists for this chain
		for _, rpc := range chainRpcs {
			if rpc.URL == request.RpcUrl {
				fmt.Printf("rpc url already exists chain_id=%s rpc_url=%s \n", request.ChainId, request.RpcUrl)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		// Add the new RPC URL
		newRpc := types.RpcInfo{URL: request.RpcUrl}
		(*rpcInfoMap)[request.ChainId] = append(chainRpcs, newRpc)

		// Write the updated map back to S3
		updatedData, err := json.Marshal(rpcInfoMap)
		if err != nil {
			fmt.Printf("failed to marshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = s3.WriteS3Object(server.USERS_BUCKET, bucketKey, updatedData)
		if err != nil {
			fmt.Printf("failed to write to s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type CustomRpcRemoveRequest struct {
	RpcUrl  string `json:"rpcUrl" validate:"required,url"`
	ChainId string `json:"chainId" validate:"required"`
}

func CreateCustomRpcRemoveQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		request := CustomRpcRemoveRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			fmt.Printf("failed to decode request err=%s \n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate the struct fields
		err = Validator.Struct(request)
		if err != nil {
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

		// Read the existing RPC info from S3
		data, err := s3.ReadS3Object(server.USERS_BUCKET, bucketKey)
		if err != nil {
			fmt.Printf("failed to read from s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rpcInfoMap := &types.RpcInfoMap{}
		err = json.Unmarshal(data, rpcInfoMap)
		if err != nil {
			fmt.Printf("failed to unmarshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check if the chain ID exists in the map
		chainRpcs, exists := (*rpcInfoMap)[request.ChainId]
		if !exists {
			fmt.Printf("chain id not found chain_id=%s \n", request.ChainId)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Find and remove the RPC URL
		found := false
		newChainRpcs := []types.RpcInfo{}
		for _, rpc := range chainRpcs {
			if rpc.URL != request.RpcUrl {
				newChainRpcs = append(newChainRpcs, rpc)
			} else {
				found = true
			}
		}

		if !found {
			fmt.Printf("rpc url not found chain_id=%s rpc_url=%s \n", request.ChainId, request.RpcUrl)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Update the map with the new list of RPCs
		(*rpcInfoMap)[request.ChainId] = newChainRpcs

		// Write the updated map back to S3
		updatedData, err := json.Marshal(rpcInfoMap)
		if err != nil {
			fmt.Printf("failed to marshal rpc info map bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = s3.WriteS3Object(server.USERS_BUCKET, bucketKey, updatedData)
		if err != nil {
			fmt.Printf("failed to write to s3 bucket_key=%s \n", bucketKey)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
