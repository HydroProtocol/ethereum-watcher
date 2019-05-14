package logic

import (
	"context"
	"ddex/watcher/models"
	"ddex/watcher/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

func FailLongWaitingCommonTransactions(ctx context.Context) {
	minString := os.Getenv("MINUTES_BEFORE_FAIL_COMMON_TRANSACTIONS")
	mins, err := strconv.ParseInt(minString, 10, 32)

	if err != nil {
		log.Fatalf("MINUTES_BEFORE_FAIL_COMMON_TRANSACTIONS missing or format error, minString: %s", minString)
	}

	for {
		txs := models.GetPendingTransactions()

		for _, tx := range txs {
			if diff := time.Now().Sub(tx.CreatedAt); diff > time.Duration(mins)*time.Minute {
				log.Infof("tx %s createdAt more than 10min, canceling...", tx.ID)
				models.UpdateCommonTransactionStatusById(tx.ID, "failed")
				utils.SendCommonTransactionWebsocketMessage(tx)
			}
		}

		if len(txs) == 0 {
			log.Infof("No pending CommonTransactions, sleep %ds", 30)
		}

		select {
		case <-ctx.Done():
			log.Info("BlockWatcher exit")
			return
		case <-time.After(30 * time.Second):
		}
	}
}
