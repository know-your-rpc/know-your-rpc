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
			err := s3.CopyS3Object(USERS_BUCKET, "public.json", USERS_BUCKET, bucketKey)
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

	rpcInfoMap := &types.RpcInfoMap{}
	err = json.Unmarshal(data, rpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain list %v", err)
	}

	rpcUrls, ok := (*rpcInfoMap)[chainId]

	if !ok {
		return nil, fmt.Errorf("couldn't find rpcUrls for chainId=%s", chainId)
	}

	return rpcUrls, nil
}
