package structs

import (
	"github.com/HydroProtocol/nights-watch/blockchain"
)

type RemovableBlock struct {
	blockchain.Block
	IsRemoved bool
}

func NewRemovableBlock(block blockchain.Block, isRemoved bool) *RemovableBlock {
	return &RemovableBlock{
		block,
		isRemoved,
	}
}

type TxAndReceipt struct {
	Tx      blockchain.Transaction
	Receipt blockchain.TransactionReceipt
}

type RemovableTxAndReceipt struct {
	*TxAndReceipt
	IsRemoved bool
	TimeStamp uint64
}

type RemovableReceiptLog struct {
	blockchain.IReceiptLog
	IsRemoved bool
}

func NewRemovableTxAndReceipt(tx blockchain.Transaction, receipt blockchain.TransactionReceipt, removed bool, timeStamp uint64) *RemovableTxAndReceipt {
	return &RemovableTxAndReceipt{
		&TxAndReceipt{
			tx,
			receipt,
		},
		removed,
		timeStamp,
	}
}

type RemovableTx struct {
	blockchain.Transaction
	IsRemoved bool
}

func NewRemovableTx(tx blockchain.Transaction, removed bool) RemovableTx {
	return RemovableTx{
		tx,
		removed,
	}
}

//
//type RemovableReceipt struct {
//	sdk.TransactionReceipt
//	IsRemoved bool
//}
//
//func NewRemovableReceipt(receipt sdk.TransactionReceipt, removed bool) RemovableReceipt {
//	return RemovableReceipt{
//		receipt,
//		removed,
//	}
//}
