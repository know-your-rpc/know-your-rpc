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
	{ChainId: "5", Name: "Goerli"},
	{ChainId: "100", Name: "Gnosis"},
	{ChainId: "137", Name: "Polygon Mainnet"},
	{ChainId: "169", Name: "Manta Pacific Mainnet"},
	{ChainId: "223", Name: "B2 Mainnet"},
	{ChainId: "252", Name: "Fraxtal"},
	{ChainId: "324", Name: "zkSync Mainnet"},
	{ChainId: "1992", Name: "Hubble Exchange"},
	{ChainId: "4200", Name: "Merlin Mainnet"},
	{ChainId: "5000", Name: "Mantle"},
	{ChainId: "7700", Name: "Canto"},
	{ChainId: "34443", Name: "Mode"},
	{ChainId: "42220", Name: "Celo Mainnet"},
	{ChainId: "42793", Name: "Etherlink Mainnet"},
	{ChainId: "48900", Name: "Zircuit Mainnet"},
	{ChainId: "60808", Name: "Bob"},
	{ChainId: "80084", Name: "Berachain bArtio"},
	{ChainId: "81457", Name: "Blast"},
	{ChainId: "111188", Name: "re.al"},
	{ChainId: "534352", Name: "Scroll"},
	{ChainId: "11155111", Name: "Sepolia"},
	{ChainId: "1329", Name: "Sei Network"},
	{ChainId: "810180", Name: "zkLink Nova Mainnet"},
	{ChainId: "7560", Name: "Cyber Mainnet"},
	{ChainId: "6001", Name: "Bounce Bit"},
	{ChainId: "2222", Name: "Kava"},
}
