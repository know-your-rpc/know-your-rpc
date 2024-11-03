package types

import "time"

type RpcInfo struct {
	URL string `json:"url"`
}

type RpcInfoMap = map[string][]RpcInfo

type Subscription struct {
	TxReceipts []TxReceipt `json:"txReceipts"`
	ValidUntil int64       `json:"validUntil"`
}

type TxReceipt struct {
	TxHash  string `json:"txHash"`
	ChainID int64  `json:"chainID"`
}

type UserData struct {
	RpcInfo      RpcInfoMap   `json:"rpcInfo"`
	Subscription Subscription `json:"subscription"`
}

func (userData *UserData) GetRpcUrlsForChainId(chainId string) ([]RpcInfo, bool) {
	rpcUrls, ok := userData.RpcInfo[chainId]
	return rpcUrls, ok
}

func (userData *UserData) IsSubscriptionValid() bool {
	return userData.Subscription.ValidUntil > time.Now().Unix()
}

func IsPublicUser(userAddress string) bool {
	return userAddress == "public"
}
