package eth_client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	transferEventSignature     = []byte("Transfer(address,address,uint256)")
	transferEventSignatureHash = crypto.Keccak256Hash(transferEventSignature)
)

type EthClient struct {
	client *ethclient.Client
}

func NewEthClient(rpcUrl string) (*EthClient, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client error=%s", err)
	}
	return &EthClient{client: client}, nil
}

type ERC20TransferEvent struct {
	Token common.Address
	From  common.Address
	To    common.Address
	Value *big.Int
}

func (e *EthClient) FetchTxReceipt(txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)

	receipt, err := e.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction receipt error=%s", err)
	}

	return receipt, nil
}

func parseERC20TransferLogs(logs []*types.Log) (*ERC20TransferEvent, error) {
	for _, log := range logs {
		if len(log.Topics) == 3 && log.Topics[0] == transferEventSignatureHash {
			return &ERC20TransferEvent{
				Token: log.Address,
				From:  common.BytesToAddress(log.Topics[1].Bytes()),
				To:    common.BytesToAddress(log.Topics[2].Bytes()),
				Value: new(big.Int).SetBytes(log.Data),
			}, nil
		}
	}
	return nil, fmt.Errorf("no transfer event found")
}

func VerifyErc20Transfer(ethClient *EthClient, txHash string, expectedFrom common.Address, expectedTo common.Address, expectedValue *big.Int, expectedToken common.Address) error {
	receipt, err := ethClient.FetchTxReceipt(txHash)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction receipt error=%s", err)
	}

	logs := receipt.Logs

	transferEvent, err := parseERC20TransferLogs(logs)
	if err != nil {
		return fmt.Errorf("failed to parse transfer event error=%s", err)
	}

	if transferEvent.From != expectedFrom {
		return fmt.Errorf("unexpected 'from' address: got %s, want %s", transferEvent.From.Hex(), expectedFrom.Hex())
	}
	if transferEvent.To != expectedTo {
		return fmt.Errorf("unexpected 'to' address: got %s, want %s", transferEvent.To.Hex(), expectedTo.Hex())
	}
	if transferEvent.Value.Cmp(expectedValue) != 0 {
		return fmt.Errorf("unexpected transfer value: got %s, want %s", transferEvent.Value.String(), expectedValue.String())
	}
	if transferEvent.Token != expectedToken {
		return fmt.Errorf("unexpected token address: got %s, want %s", transferEvent.Token.Hex(), expectedToken.Hex())
	}

	return nil
}
