package plugin

import (
	"github.com/HydroProtocol/nights-watch/ethrpc"
	"github.com/HydroProtocol/nights-watch/utils"
	"github.com/shopspring/decimal"
)

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

type ERC20TransferPlugin struct {
	callback func(tokenAddress, from, to string, amount decimal.Decimal)
}

func NewERC20TransferPlugin(callback func(tokenAddress, from, to string, amount decimal.Decimal)) *ERC20TransferPlugin {
	return &ERC20TransferPlugin{callback}
}

func (p *ERC20TransferPlugin) Accept(tx *ethrpc.Transaction, receipt *ethrpc.TransactionReceipt) {
	if p.callback != nil {
		events := extractERC20TransfersIfExist(receipt)

		for _, e := range events {
			p.callback(e.token, e.from, e.to, e.value)
		}
	}
}

type TransferEvent struct {
	token string
	from  string
	to    string
	value decimal.Decimal
}

func extractERC20TransfersIfExist(receipt *ethrpc.TransactionReceipt) (rst []TransferEvent) {
	transferEventSig := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	for _, log := range receipt.Logs {
		if len(log.Topics) == 3 && log.Topics[0] == transferEventSig {
			from := log.Topics[1]
			to := log.Topics[2]

			amount, ok := utils.HexToDecimal(log.Data)

			if ok {
				rst = append(rst, TransferEvent{log.Address, from, to, amount})
			}
		}
	}

	return
}
