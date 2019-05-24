package rpc

import (
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/onrik/ethrpc"
)

type EthBlockChainRPC struct {
	rpcImpl *ethrpc.EthRPC
}

func NewEthRPC(api string) *EthBlockChainRPC {
	rpc := ethrpc.New(api)

	return &EthBlockChainRPC{rpc}
}

func (rpc EthBlockChainRPC) GetBlockByNum(num uint64) (sdk.Block, error) {
	b, err := rpc.rpcImpl.EthGetBlockByNumber(int(num), true)
	return &ethereum.EthereumBlock{b}, err
}

func (rpc EthBlockChainRPC) GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error) {
	receipt, error := rpc.rpcImpl.EthGetTransactionReceipt(txHash)
	return &ethereum.EthereumTransactionReceipt{receipt}, error
}

func (rpc EthBlockChainRPC) GetCurrentBlockNum() (uint64, error) {
	num, err := rpc.rpcImpl.EthBlockNumber()
	return uint64(num), err
}
