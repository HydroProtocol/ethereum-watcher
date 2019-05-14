package plugin

import "github.com/HydroProtocol/nights-watch/ethrpc"

type ITxReceiptPlugin interface {
	Accept(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt)
}

type TxReceiptPlugin struct {
	callback func(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt)
}

func NewTxReceiptPlugin(callback func(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt)) *TxReceiptPlugin {
	return &TxReceiptPlugin{callback}
}

func (p TxReceiptPlugin) Accept(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt) {
	if p.callback != nil {
		p.callback(tx, receipt)
	}
}
