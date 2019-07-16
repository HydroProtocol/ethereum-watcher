package structs

import "github.com/HydroProtocol/hydro-sdk-backend/sdk"

type RemovableBlock struct {
	sdk.Block
	IsRemoved bool
}

func NewRemovableBlock(block sdk.Block, isRemoved bool) *RemovableBlock {
	return &RemovableBlock{
		block,
		isRemoved,
	}
}

type TxAndReceipt struct {
	Tx      sdk.Transaction
	Receipt sdk.TransactionReceipt
}

type RemovableTxAndReceipt struct {
	*TxAndReceipt
	IsRemoved bool
	TimeStamp uint64
}

type RemovableReceiptLog struct {
	sdk.IReceiptLog
	sdk.Block
	IsRemoved bool
}

func NewRemovableTxAndReceipt(tx sdk.Transaction, receipt sdk.TransactionReceipt, removed bool, timeStamp uint64) *RemovableTxAndReceipt {
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
	sdk.Transaction
	IsRemoved bool
}

func NewRemovableTx(tx sdk.Transaction, removed bool) RemovableTx {
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
