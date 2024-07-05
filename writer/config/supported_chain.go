package config

type ChainConfig struct {
	ChainId string `json:"chainId"`
	Name    string `json:"name"`
}

var SUPPORTED_CHAINS = []ChainConfig{
	{ChainId: "1", Name: "Ethereum"},
	{ChainId: "56", Name: "BNB Smart Chain Mainnet"},
	{ChainId: "42161", Name: "Arbitrum One"},
	{ChainId: "8453", Name: "Base"},
	{ChainId: "43114", Name: "Avalanche C-Chain"},
	{ChainId: "59144", Name: "Linea"},
	{ChainId: "137", Name: "Polygon Mainnet"},
	{ChainId: "10", Name: "OP Mainnet"},
}
