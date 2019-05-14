package plugin

import "github.com/HydroProtocol/nights-watch/ethrpc"

type ITxPlugin interface {
	AcceptTx(transaction ethrpc.Transaction)
}

type TxHashPlugin struct {
	callback func(txHash string, isRemoved bool)
}

func (p TxHashPlugin) AcceptTx(transaction ethrpc.Transaction) {
	if p.callback != nil {
		p.callback(transaction.Hash, transaction.IsRemoved)
	}
}

func NewTxHashPlugin(callback func(txHash string, isRemoved bool)) TxHashPlugin {
	return TxHashPlugin{
		callback: callback,
	}
}
