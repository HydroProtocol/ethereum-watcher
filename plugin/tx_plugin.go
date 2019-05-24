package plugin

import (
	"github.com/HydroProtocol/nights-watch/structs"
)

type ITxPlugin interface {
	AcceptTx(transaction structs.RemovableTx)
}

type TxHashPlugin struct {
	callback func(txHash string, isRemoved bool)
}

func (p TxHashPlugin) AcceptTx(transaction structs.RemovableTx) {
	if p.callback != nil {
		p.callback(transaction.GetHash(), transaction.IsRemoved)
	}
}

func NewTxHashPlugin(callback func(txHash string, isRemoved bool)) TxHashPlugin {
	return TxHashPlugin{
		callback: callback,
	}
}
