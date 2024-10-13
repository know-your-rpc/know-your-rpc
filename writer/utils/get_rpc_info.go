package utils

import (
	"encoding/json"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
	"log"
	"sync"
	"time"
)

type ChainRpcInfoReader struct {
	cachedRpcInfoMutex sync.RWMutex
	cachedRpcInfo      *types.RpcInfoMap
	eTagCache          map[string]*types.UserData
	interval           time.Duration
}

func CreateChainRpcInfoReader(interval time.Duration) *ChainRpcInfoReader {
	return &ChainRpcInfoReader{
		cachedRpcInfoMutex: sync.RWMutex{},
		eTagCache:          make(map[string]*types.UserData),
		interval:           interval,
		cachedRpcInfo:      nil,
	}
}

func (c *ChainRpcInfoReader) Start() {
	go func() {
		for {
			rpcInfo, err := c.UpdateRpcInfo()
			if err != nil {
				log.Printf("failed to update rpc info: %v", err)
			} else {
				c.cachedRpcInfoMutex.Lock()
				c.cachedRpcInfo = rpcInfo
				c.cachedRpcInfoMutex.Unlock()
			}
			time.Sleep(c.interval)
		}
	}()
}

func (c *ChainRpcInfoReader) GetRpcInfo() (types.RpcInfoMap, error) {
	c.cachedRpcInfoMutex.RLock()
	defer c.cachedRpcInfoMutex.RUnlock()

	if c.cachedRpcInfo == nil {
		return nil, fmt.Errorf("cached rpc info is nil")
	}
	return *c.cachedRpcInfo, nil
}

func (c *ChainRpcInfoReader) UpdateRpcInfo() (*types.RpcInfoMap, error) {
	rpcInfoMap := make(types.RpcInfoMap)

	userBuckets, err := s3.ListS3Objects(server.USERS_BUCKET)
	if err != nil {
		return nil, fmt.Errorf("failed to list user buckets: %w", err)
	}

	for _, bucket := range userBuckets {
		bucketRpcMap, ok := c.eTagCache[*bucket.Key+"-"+*bucket.ETag]

		if !ok {
			fmt.Println("READING FROM S3")
			bucketContent, err := s3.ReadS3Object(server.USERS_BUCKET, *bucket.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to read bucket %s: %w", *bucket.Key, err)
			}

			err = json.Unmarshal(bucketContent, &bucketRpcMap)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal bucket content %s: %w", *bucket.Key, err)
			}

			c.eTagCache[*bucket.Key+"-"+*bucket.ETag] = bucketRpcMap
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

func containsRPC(rpcs []types.RpcInfo, rpc types.RpcInfo) bool {
	for _, r := range rpcs {
		if r.URL == rpc.URL {
			return true
		}
	}
	return false
}
