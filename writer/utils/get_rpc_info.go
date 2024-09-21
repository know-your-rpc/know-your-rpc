package utils

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
)

const (
	CHAIN_LIST_FILE = "chain-list.json"
)

func containsRPC(rpcs []types.RpcInfo, rpc types.RpcInfo) bool {
	for _, r := range rpcs {
		if r.URL == rpc.URL {
			return true
		}
	}
	return false
}

// TODO: optimize by checking eTag of file in bucket - we will have to hold eTag in mem
// TODO: read on every n-th iteration
func ReadRpcInfo() (*types.RpcInfoMap, error) {

	rpcInfoMap := make(types.RpcInfoMap)

	userBuckets, err := s3.ListS3Objects(server.USERS_BUCKET)
	if err != nil {
		return nil, fmt.Errorf("failed to list user buckets: %w", err)
	}

	for _, bucket := range userBuckets {

		bucketContent, err := s3.ReadS3Object(server.USERS_BUCKET, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to read bucket %s: %w", bucket, err)
		}

		var bucketRpcMap types.UserStore
		err = json.Unmarshal(bucketContent, &bucketRpcMap)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal bucket content %s: %w", bucket, err)
		}

		for chainID, rpcs := range bucketRpcMap.RpcInfo {
			if _, exists := rpcInfoMap[chainID]; !exists {
				rpcInfoMap[chainID] = []types.RpcInfo{}
			}

			for _, rpc := range rpcs {
				if !containsRPC(rpcInfoMap[chainID], rpc) {
					rpcInfoMap[chainID] = append(rpcInfoMap[chainID], rpc)
				}
			}
		}
	}

	return &rpcInfoMap, nil
}
