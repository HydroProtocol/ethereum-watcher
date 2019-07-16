package rpc

import (
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"time"
)

type EthBlockChainRPCWithRetry struct {
	*EthBlockChainRPC
	maxRetryTimes int
}

func NewEthRPCWithRetry(api string, maxRetryCount int) *EthBlockChainRPCWithRetry {
	rpc := NewEthRPC(api)

	return &EthBlockChainRPCWithRetry{rpc, maxRetryCount}
}

func (rpc EthBlockChainRPCWithRetry) GetBlockByNum(num uint64) (rst sdk.Block, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetBlockByNum(num)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}

func (rpc EthBlockChainRPCWithRetry) GetTransactionReceipt(txHash string) (rst sdk.TransactionReceipt, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetTransactionReceipt(txHash)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}

func (rpc EthBlockChainRPCWithRetry) GetCurrentBlockNum() (rst uint64, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetCurrentBlockNum()
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}
func (rpc EthBlockChainRPCWithRetry) GetLogs(
	fromBlockNum, toBlockNum uint64,
	address string,
	topics []string,
) (rst []sdk.IReceiptLog, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetLogs(fromBlockNum, toBlockNum, address, topics)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}
