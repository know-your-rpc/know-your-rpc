package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const USERS_BUCKET = "know-your-rpc-users"
const PUBLIC_S3_KEY = "public.json"

func ReadRpcUrlsForUser(userAddress string, chainId string) ([]types.RpcInfo, error) {
	bucketKey := fmt.Sprintf("%s.json", userAddress)

	data, err := s3.ReadS3Object(USERS_BUCKET, bucketKey)

	if err != nil {
		var noSuchKeyError *s3Types.NoSuchKey
		if errors.As(err, &noSuchKeyError) {
			fmt.Printf("copying public.json to %s, because it was not created yet", bucketKey)
			err := s3.CopyS3Object(USERS_BUCKET, PUBLIC_S3_KEY, USERS_BUCKET, bucketKey)
			if err != nil {
				return nil, fmt.Errorf("failed to copy s3 object error=%s", err)
			}
			data, err = s3.ReadS3Object(USERS_BUCKET, bucketKey)
			if err != nil {
				return nil, fmt.Errorf("failed to read from s3 after copying it bucket=%s bucketKey=%s err=%s", USERS_BUCKET, bucketKey, err)
			}
		} else {
			return nil, fmt.Errorf("failed to read from s3 bucket=%s bucketKey=%s err=%s", USERS_BUCKET, bucketKey, err)
		}
	}

	publicData, err := s3.ReadS3Object(USERS_BUCKET, PUBLIC_S3_KEY)
	if err != nil {
		return nil, fmt.Errorf("failed to read public RPC info: %v", err)
	}

	publicRpcInfoMap := &types.RpcInfoMap{}
	err = json.Unmarshal(publicData, publicRpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public RPC info: %v", err)
	}

	userRpcInfoMap := &types.RpcInfoMap{}
	err = json.Unmarshal(data, userRpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user's RPC info: %v", err)
	}

	//TODO: this is happening many times because server is sending multiple requests to the same user, find a way to avoid it
	// Check for missing chainIds in user's data
	updated := false
	for chainId, rpcInfos := range *publicRpcInfoMap {
		if _, exists := (*userRpcInfoMap)[chainId]; !exists {
			(*userRpcInfoMap)[chainId] = rpcInfos
			updated = true
		}
	}

	// If updates were made, save back to S3
	if updated {
		fmt.Printf("updating user's RPC info for userAddress=%s chainId=%s\n", userAddress, chainId)
		updatedData, err := json.Marshal(userRpcInfoMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal updated RPC info: %v", err)
		}

		err = s3.WriteS3Object(USERS_BUCKET, bucketKey, updatedData)
		if err != nil {
			return nil, fmt.Errorf("failed to write updated RPC info to S3: %v", err)
		}
	}

	rpcInfoMap := &types.RpcInfoMap{}
	err = json.Unmarshal(data, rpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain list %v", err)
	}

	rpcUrls, ok := (*userRpcInfoMap)[chainId]

	if !ok {
		return nil, fmt.Errorf("couldn't find rpcUrls for chainId=%s", chainId)
	}

	return rpcUrls, nil
}
