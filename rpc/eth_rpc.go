package rpc

import (
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"strconv"
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

func (rpc EthBlockChainRPC) GetLogs(
	fromBlockNum, toBlockNum uint64,
	address string,
	topics []string,
) ([]sdk.IReceiptLog, error) {

	logs, err := rpc.rpcImpl.EthGetLogs(ethrpc.FilterParams{
		FromBlock: "0x" + strconv.FormatUint(fromBlockNum, 16),
		ToBlock:   "0x" + strconv.FormatUint(toBlockNum, 16),
		Address:   []string{address},
		Topics:    [][]string{topics},
	})
	if err != nil {
		fmt.Println("EthGetLogs err:", err)
		return nil, err
	}

	logrus.Debugf("EthGetLogs logs count at block(%d - %d): %d", fromBlockNum, toBlockNum, len(logs))

	var result []sdk.IReceiptLog
	for _, l := range logs {
		logrus.Debugf("EthGetLogs receipt log: %+v", l)

		result = append(result, ethereum.ReceiptLog{Log: &l})
	}

	return result, err
}
