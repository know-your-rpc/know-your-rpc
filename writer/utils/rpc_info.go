package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type RpcInfo struct {
	URL string `json:"url"`
}

type RpcInfoMap = map[string][]RpcInfo

const (
	CHAIN_LIST_FILE = "chain-list.json"
)

func ReadRpcInfo() (*RpcInfoMap, error) {
	body, err := os.ReadFile(CHAIN_LIST_FILE)

	if err != nil {
		return nil, fmt.Errorf("failed to read chain list %v", err)
	}

	rpcInfoMap := &RpcInfoMap{}
	err = json.Unmarshal(body, rpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain list %v", err)
	}

	return rpcInfoMap, nil
}
