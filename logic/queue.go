package logic

import (
	"ddex/watcher/connection"
	"ddex/watcher/models"
	"encoding/json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"time"
)

type QueueTransactionMessage struct {
	Pair                  string          `json:"pair"`
	EventType             string          `json:"eventType"`
	TransactionID         string          `json:"transactionId"`
	GasPrice              decimal.Decimal `json:"gasPrice"`
	GasLimit              decimal.Decimal `json:"gasLimit"`
	GasUsed               decimal.Decimal `json:"gasUsed"`
	ExecutedAt            time.Time       `json:"executedAt"`
	Nonce                 int             `json:"nonce"`
	ConfirmedBlockNumber  int             `json:"confirmedBlockNumber"`
}

func enqueueTransaction(ct *models.CommonTransaction, pair string) {
	var message *QueueTransactionMessage

	if ct.Status == "successful" {
		message = &QueueTransactionMessage{
			Pair:          pair,
			EventType:     "event/EVENT_TX_SUCCESS",
			TransactionID: ct.ID,
			GasLimit:      ct.GasLimit,
			GasPrice:      ct.GasPrice,
			GasUsed:       ct.GasUsed,
			ExecutedAt:    ct.ExecutedAt,
			Nonce:         ct.Nonce,
		}
	} else if ct.Status == "failed" {
		message = &QueueTransactionMessage{
			Pair:          pair,
			EventType:     "event/EVENT_TX_FAIL",
			TransactionID: ct.ID,
			GasLimit:      ct.GasLimit,
			GasPrice:      ct.GasPrice,
			GasUsed:       ct.GasUsed,
			ExecutedAt:    ct.ExecutedAt,
			Nonce:         ct.Nonce,
		}
	} else if ct.Status == "discarded" {
		message = &QueueTransactionMessage{
			Pair:          pair,
			EventType: "event/EVENT_TX_DISCARD",
		}
	}

	jsonString, err := json.Marshal(message)

	if err != nil {
		panic(err)
	}

	log.Info("enqueue message ", string(jsonString))
	connection.Redis.LPush("ENGINE_EVENT_QUEUE:" + pair, string(jsonString))
}
