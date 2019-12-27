package plugin

import (
	"github.com/HydroProtocol/nights-watch/structs"
	"strings"
)

type IReceiptLogPlugin interface {
	FromContract() string
	InterestedTopics() []string
	NeedReceiptLog(receiptLog *structs.RemovableReceiptLog) bool
	Accept(receiptLog *structs.RemovableReceiptLog)
}

type ReceiptLogPlugin struct {
	contract string
	topics   []string
	callback func(receiptLog *structs.RemovableReceiptLog)
}

func NewReceiptLogPlugin(
	contract string,
	topics []string,
	callback func(receiptLog *structs.RemovableReceiptLog),
) *ReceiptLogPlugin {
	return &ReceiptLogPlugin{
		contract: contract,
		topics:   topics,
		callback: callback,
	}
}

func (p *ReceiptLogPlugin) FromContract() string {
	return p.contract
}

func (p *ReceiptLogPlugin) InterestedTopics() []string {
	return p.topics
}

func (p *ReceiptLogPlugin) Accept(receiptLog *structs.RemovableReceiptLog) {
	if p.callback != nil {
		p.callback(receiptLog)
	}
}

// simplified version of specifying topic filters
// https://github.com/ethereum/wiki/wiki/JSON-RPC#a-note-on-specifying-topic-filters
func (p *ReceiptLogPlugin) NeedReceiptLog(receiptLog *structs.RemovableReceiptLog) bool {
	contract := receiptLog.GetAddress()
	if strings.ToLower(p.contract) != strings.ToLower(contract) {
		return false
	}

	var firstTopic string
	if len(receiptLog.GetTopics()) > 0 {
		firstTopic = receiptLog.GetTopics()[0]
	}

	for _, interestedTopic := range p.topics {
		if strings.ToLower(firstTopic) == strings.ToLower(interestedTopic) {
			return true
		}
	}

	return false
}
