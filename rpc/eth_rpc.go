package rpc

import (
	"errors"
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
	return rpc.getBlockByNum(num, true)
}

func (rpc EthBlockChainRPC) GetLiteBlockByNum(num uint64) (sdk.Block, error) {
	return rpc.getBlockByNum(num, false)
}

func (rpc EthBlockChainRPC) getBlockByNum(num uint64, withTx bool) (sdk.Block, error) {
	b, err := rpc.rpcImpl.EthGetBlockByNumber(int(num), withTx)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, errors.New("nil block")
	}

	return &ethereum.EthereumBlock{b}, err
}

func (rpc EthBlockChainRPC) GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error) {
	receipt, err := rpc.rpcImpl.EthGetTransactionReceipt(txHash)
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("nil receipt")
	}

	return &ethereum.EthereumTransactionReceipt{receipt}, err
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

	var result []sdk.IReceiptLog
	for i := 0; i < len(logs); i++ {
		l := logs[i]

		logrus.Debugf("EthGetLogs receipt log: %+v", l)

		result = append(result, ethereum.ReceiptLog{Log: &l})
	}

	return result, err
}
