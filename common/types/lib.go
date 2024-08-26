package types

type RpcInfo struct {
	URL string `json:"url"`
}

type RpcInfoMap = map[string][]RpcInfo
