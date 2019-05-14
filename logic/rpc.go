package logic

import (
	"ddex/watcher/connection"
	"github.com/onrik/ethrpc"
	"time"
)

func getCurrentBlockNumber() int64 {
	blockNumber, err := connection.RPC.EthBlockNumber()

	if err != nil {
		panic(err)
	}

	return int64(blockNumber)
}

func getBlockByNumber(number int64) *ethrpc.Block {
	var err error

	for i := 0; i < 5; i++ {
		block, err := connection.RPC.EthGetBlockByNumber(int(number), true)

		if err != nil {
			println("get Block by Number failed, Retry", i, number, err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}

		return block
	}

	println("get Block by Number failed", number, err)
	panic(err)
}

func getTransactionReceipt(hash string) *ethrpc.TransactionReceipt {
	tx, err := connection.RPC.EthGetTransactionReceipt(hash)

	if err != nil {
		return nil
	}

	// ropsten doesn't return nil
	if tx.TransactionHash == "" {
		return nil
	}

	return tx
}

func getTransaction(hash string) *ethrpc.Transaction {
	tx, err := connection.RPC.EthGetTransactionByHash(hash)

	if err != nil {
		return nil
	}

	return tx
}
