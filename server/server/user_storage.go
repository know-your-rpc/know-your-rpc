package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"koonopek/know_your_rpc/common/s3"
	"koonopek/know_your_rpc/common/types"
	"sync"
	"time"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const USERS_BUCKET = "know-your-rpc-users"
const PUBLIC_S3_KEY = "public.json"

// Add these new structures and variables
type userDataCacheEntry struct {
	data      *types.UserStore
	expiresAt time.Time
}

var (
	userDataCache      = make(map[string]userDataCacheEntry)
	userDataCacheMutex sync.RWMutex
	userDataCacheTTL   = 1 * time.Hour
)

func ReadAndUpdateRpcUrlsForUserAndChainId(userAddress string, chainId string) ([]types.RpcInfo, error) {
	privateUserStore, err := ReadAndUpdateUserData(userAddress)

	if err != nil {
		return nil, fmt.Errorf("couldn't fetch user data userAddress=%s", userAddress)
	}

	rpcUrls, ok := privateUserStore.RpcInfo[chainId]

	if !ok {
		return nil, fmt.Errorf("couldn't find rpcUrls for chainId=%s", chainId)
	}

	return rpcUrls, nil
}

func ReadAndUpdateUserData(userAddress string) (*types.UserStore, error) {
	// Check cache first
	userDataCacheMutex.RLock()
	if entry, found := userDataCache[userAddress]; found && time.Now().Before(entry.expiresAt) {
		userDataCacheMutex.RUnlock()
		return entry.data, nil
	}
	userDataCacheMutex.RUnlock()

	// If not in cache or expired, fetch from S3
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

	publicUserStore := &types.UserStore{}
	err = json.Unmarshal(publicData, publicUserStore)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public RPC info: %v", err)
	}

	privateUserStore := &types.UserStore{}
	err = json.Unmarshal(data, privateUserStore)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user's RPC info: %v", err)
	}

	// Check for missing chainIds in user's data
	updated := false
	for chainId, rpcInfos := range publicUserStore.RpcInfo {
		if _, exists := privateUserStore.RpcInfo[chainId]; !exists {
			privateUserStore.RpcInfo[chainId] = rpcInfos
			updated = true
		}
	}

	// If updates were made, save back to S3
	if updated {
		fmt.Printf("updating user's RPC info for userAddress=%s \n", userAddress)
		updatedData, err := json.Marshal(privateUserStore)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal updated RPC info: %v", err)
		}

		err = s3.WriteS3Object(USERS_BUCKET, bucketKey, updatedData)
		if err != nil {
			return nil, fmt.Errorf("failed to write updated RPC info to S3: %v", err)
		}
	}

	// Update cache
	userDataCacheMutex.Lock()
	userDataCache[userAddress] = userDataCacheEntry{
		data:      privateUserStore,
		expiresAt: time.Now().Add(userDataCacheTTL),
	}
	userDataCacheMutex.Unlock()

	return privateUserStore, nil
}
