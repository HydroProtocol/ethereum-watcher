package rpc

import (
	"errors"
	"strconv"

	"github.com/onrik/ethrpc"
	"github.com/rakshasa/ethereum-watcher/blockchain"
	"github.com/sirupsen/logrus"
)

type EthBlockChainRPC struct {
	rpcImpl *ethrpc.EthRPC
}

func NewEthRPC(api string, options ...func(rpc *ethrpc.EthRPC)) *EthBlockChainRPC {
	rpc := ethrpc.New(api, options...)

	return &EthBlockChainRPC{rpc}
}

func (rpc EthBlockChainRPC) GetBlockByNum(num uint64) (blockchain.Block, error) {
	return rpc.getBlockByNum(num, true)
}

func (rpc EthBlockChainRPC) GetLiteBlockByNum(num uint64) (blockchain.Block, error) {
	return rpc.getBlockByNum(num, false)
}

func (rpc EthBlockChainRPC) getBlockByNum(num uint64, withTx bool) (blockchain.Block, error) {
	b, err := rpc.rpcImpl.EthGetBlockByNumber(int(num), withTx)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errors.New("nil block")
	}

	return &blockchain.EthereumBlock{b}, err
}

func (rpc EthBlockChainRPC) GetTransactionReceipt(txHash string) (blockchain.TransactionReceipt, error) {
	receipt, err := rpc.rpcImpl.EthGetTransactionReceipt(txHash)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("nil receipt")
	}

	return &blockchain.EthereumTransactionReceipt{receipt}, err
}

func (rpc EthBlockChainRPC) GetCurrentBlockNum() (uint64, error) {
	num, err := rpc.rpcImpl.EthBlockNumber()
	return uint64(num), err
}

func (rpc EthBlockChainRPC) GetLogs(
	fromBlockNum, toBlockNum uint64,
	address string,
	topics []string,
) ([]blockchain.IReceiptLog, error) {

	filterParam := ethrpc.FilterParams{
		FromBlock: "0x" + strconv.FormatUint(fromBlockNum, 16),
		ToBlock:   "0x" + strconv.FormatUint(toBlockNum, 16),
		Address:   []string{address},
		Topics:    [][]string{topics},
	}

	logs, err := rpc.rpcImpl.EthGetLogs(filterParam)
	if err != nil {
		logrus.Warnf("EthGetLogs err: %s, params: %+v", err, filterParam)
		return nil, err
	}

	logrus.Debugf("EthGetLogs logs count at block(%d - %d): %d", fromBlockNum, toBlockNum, len(logs))

	var result []blockchain.IReceiptLog
	for i := 0; i < len(logs); i++ {
		l := logs[i]

		logrus.Debugf("EthGetLogs receipt log: %+v", l)

		result = append(result, blockchain.ReceiptLog{Log: &l})
	}

	return result, err
}
