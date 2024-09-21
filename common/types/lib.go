package types

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

type UserStore struct {
	RpcInfo       RpcInfoMap   `json:"rpcInfo"`
	Subscriptions Subscription `json:"subscriptions"`
}
