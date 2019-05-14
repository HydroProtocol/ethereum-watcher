package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/nights-watch/logger"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var confirmBlockOffset int64

func init() {
	offset := os.Getenv("CONFIRM_BLOCK_OFFSET")
	logger.BlockWatcherLogger.Infof("Set confirm block offset %s", offset)
	confirmBlockOffset, _ = strconv.ParseInt(offset, 10, 64)
}

func getTransactionInfoAndReceipt(onChainHash string) (*ethrpc.Transaction, *ethrpc.TransactionReceipt) {
	logger.BlockWatcherLogger.Debugf("Request tx %s info and receipt...", onChainHash)
	txReceiptChannel := make(chan *ethrpc.TransactionReceipt)

	go func() {
		txReceiptChannel <- getTransactionReceipt(onChainHash)
	}()

	txInfoChannel := make(chan *ethrpc.Transaction)
	go func() {
		txInfoChannel <- getTransaction(onChainHash)
	}()

	logger.BlockWatcherLogger.Debugf("Request tx %s info and receipt done", onChainHash)
	return <-txInfoChannel, <-txReceiptChannel
}

func getTransactionStatusByReceiptStatus(status int) string {
	if status == 1 {
		return "successful"
	} else if status == 0 {
		return "failed"
	} else {
		panic(fmt.Sprintf("wrong tx receipt status %d", status))
	}
}

/*
     pending -> successful
     pending -> failed
      failed -> successful (ignore for now)
  successful -> failed     (ignore for now)
*/
func handleTransaction(tx *models.LaunchLog, block *ethrpc.Block, confirm bool) {
	startTime := time.Now()
	defer ddexUtils.Monitor.MonitorTime("watch_business_transaction", float64(time.Since(startTime)) / 1000000)

	logger.BlockWatcherLogger.Debugf("HandleTransaction(confirm: %t) %v", confirm, tx.Hash)

	txInfo, txReceipt := getTransactionInfoAndReceipt(tx.Hash.String)

	if txInfo == nil {
		logger.BlockWatcherLogger.Debugf("HandleTransaction(confirm: %t) %v info not found\n", confirm, tx.Hash)
		ddexUtils.Monitor.MonitorCount("watch_business_transaction_tx_info_failed")
		return
	}

	if txReceipt == nil {
		logger.BlockWatcherLogger.Debugf("HandleTransaction(confirm: %t) %v receipt not found\n", confirm, tx.Hash)
		ddexUtils.Monitor.MonitorCount("watch_business_transaction_tx_receipt_failed")
		return
	}

	logger.BlockWatcherLogger.WithFields(log.Fields{
		"ID":              tx.Hash,
		"confirm":         confirm,
		"txModelStatus":   tx.Status,
		"txReceiptStatus": txReceipt.Status,
	}).Info("handleTransaction")

	status := getTransactionStatusByReceiptStatus(txReceipt.Status)

	if strings.EqualFold("approve", tx.Itype.String) || strings.EqualFold("withdraw", tx.Itype.String) || strings.EqualFold("transfer", tx.Itype.String) {
		if strings.EqualFold("successful", status) {
			models.UpdateLaunchStatusAndIgnoreOtherSameNonceLogs(tx, "success")
		} else {
			models.UpdateLaunchStatusAndIgnoreOtherSameNonceLogs(tx, "failed")
		}

		return
	}

	/*
		pending -> successful
		pending -> failed
	*/
	if tx.Status.String == "pending" {
		enqueueTransaction(&models.CommonTransaction{
			ID:         tx.Hash.String,
			Status:     status,
			GasLimit:   decimal.New(int64(txInfo.Gas), 0),
			GasPrice:   decimal.NewFromBigInt(&txInfo.GasPrice, 0),
			GasUsed:    decimal.New(int64(txReceipt.GasUsed), 0),
			Nonce:      txInfo.Nonce,
			ExecutedAt: time.Unix(int64(block.Timestamp), 0),
		}, tx.Pair.String)
	}

	if confirm {
		models.UpdateLaunchLogConfirmedBlockNumberByOnChainHash(tx.Hash.String, block.Number)
	}
}

/*
     pending -> successful
     pending -> failed
      failed -> successful
  successful -> failed
*/
func handleCommonTransaction(tx *models.CommonTransaction, block *ethrpc.Block, confirm bool) {
	startTime := time.Now()
	defer ddexUtils.Monitor.MonitorTime("watch_common_transaction", float64(time.Since(startTime)) / 1000000)

	logger.BlockWatcherLogger.Debugf("HandleCommonTransaction %s", tx.ID)

	// Do not check pending here, make sure wrong status commonTransaction can be fixed by admin reload

	txInfo, txReceipt := getTransactionInfoAndReceipt(tx.ID)

	if txInfo == nil {
		logger.BlockWatcherLogger.Debugf("HandleCommonTransaction %s info not found\n", tx.ID)
		ddexUtils.Monitor.MonitorCount("watch_common_transaction_tx_info_failed")
		return
	}

	if txReceipt == nil {
		logger.BlockWatcherLogger.Debugf("HandleCommonTransaction %s receipt not found\n", tx.ID)
		ddexUtils.Monitor.MonitorCount("watch_common_transaction_tx_receipt_failed")
		return
	}

	status := getTransactionStatusByReceiptStatus(txReceipt.Status)

	logger.BlockWatcherLogger.WithFields(log.Fields{
		"ID":              tx.ID,
		"confirm":         confirm,
		"txModelStatus":   tx.Status,
		"txReceiptStatus": txReceipt.Status,
	}).Info("handleCommonTransaction")

	oldStatus := tx.Status
	tx.Status = status

	if oldStatus == "pending" {
		if status == "successful" {
			// send mission kafka message
			utils.SendMissionMessage(tx)
		}

		// v2 ws notification
		utils.SendCommonTransactionWebsocketMessage(tx)

		// v1 ws notification
		txJSON, _ := json.Marshal(tx)
		var tmp map[string]interface{}
		json.Unmarshal(txJSON, &tmp)
		connection.Emitter.EmitTo([]string{fmt.Sprintf("account:%s", tx.Account)}, "onCommonTransaction", tmp)
	}

	if confirm {
		models.ConfirmUpdateCommonTransactionByID(tx.ID, status, txInfo.GasPrice, txInfo.Gas, txReceipt.GasUsed, txInfo.Nonce, txReceipt.BlockNumber)
	} else {
		models.UpdateCommonTransactionByID(tx.ID, status, txInfo.GasPrice, txInfo.Gas, txReceipt.GasUsed, txInfo.Nonce)
	}
}

func syncTransaction(ID string, block *ethrpc.Block, confirm bool) {
	startTime := time.Now()
	defer ddexUtils.Monitor.MonitorTime("watch_transaction", float64(time.Since(startTime)) / 1000000)

	commonTx := models.GetCommonTransactionByID(ID)

	if commonTx != nil {
		handleCommonTransaction(commonTx, block, confirm)
		return
	}

	tx := models.GetLaunchLogByOnChainID(ID)
	if tx != nil {
		handleTransaction(tx, block, confirm)
		return
	}

	// logger.BlockWatcherLogger.Debugf("Skip %s, no records in our system", ID)
	return
}

func syncBlock(number int64, confirm bool) {
	if number <= 0 {
		return
	}

	startTime := time.Now()
	if confirm {
		logger.BlockWatcherLogger.Printf("Syncing block %d for confirmation", number)
	} else {
		logger.BlockWatcherLogger.Printf("Syncing block %d", number)
	}

	block := getBlockByNumber(number)

	var wg sync.WaitGroup

	for _, t := range block.Transactions {
		wg.Add(1)
		go func(ID string) {
			defer wg.Done()
			syncTransaction(ID, block, confirm)
		}(t.Hash)
	}

	wg.Wait()
	ddexUtils.Monitor.MonitorTime("watch_block", float64(time.Since(startTime)) / 1000000)
}

func StartBlockWatcher(cxt context.Context) {
	for {
		// graceful exit
		select {
		case <-cxt.Done():
			logger.BlockWatcherLogger.Info("BlockWatcher exit")
			return
		default:
		}

		lastBlockNumber := getLastSyncedBlockNumber()
		currentBlockNumber := getCurrentBlockNumber()

		ddexUtils.Monitor.MonitorValue("watcher_last_sync_block_number", float64(lastBlockNumber))
		if currentBlockNumber <= lastBlockNumber {
			time.Sleep(3 * time.Second)
			continue
		}

		var nextNumber int64
		if lastBlockNumber == -1 {
			nextNumber = currentBlockNumber
		} else {
			nextNumber = lastBlockNumber + 1
		}

		logger.BlockWatcherLogger.Printf("Catch up: current %d, %d synced, offset: %d", currentBlockNumber, lastBlockNumber, currentBlockNumber-lastBlockNumber)
		syncBlock(nextNumber, false)
		syncBlock(nextNumber-confirmBlockOffset, true)
		saveSyncedBlockNumber(nextNumber)
	}
}

// HandleReloadRequest ..
func HandleReloadRequest(ctx context.Context) {
	for {
		// graceful exit
		select {
		case <-ctx.Done():
			logger.ReloadLogger.Info("Transaction Reloader exit")
			return
		default:
		}

		result := connection.Redis.BRPop(0, "watcher/WATCHER_QUEUE")
		jsonString := result.Val()[1]
		var message map[string]string

		err := json.Unmarshal([]byte(jsonString), &message)

		if err != nil {
			logger.ReloadLogger.Error("parse json error", err)
		}

		switch message["eventType"] {
		case "watcher/event/EVENT_RECHECK_TRANSACTION":
			logger.ReloadLogger.Debugf("received Message %s", message)

			txID := message["transactionId"]
			launchlogs := models.GetLaunchLogByTransactionId(txID)
			var txReceipt *ethrpc.TransactionReceipt
			for _, launchlog := range launchlogs {
				txReceipt = getTransactionReceipt(launchlog.Hash.String)
				if txReceipt != nil {
					break
				}
		  }

			if txReceipt == nil {
				logger.ReloadLogger.Debugf("txID: %s not exist", txID)
				break
			} else {
				logger.ReloadLogger.Debugf("txReceipt %+v", txReceipt)
			}

			block := getBlockByNumber(int64(txReceipt.BlockNumber))
			currentBlockNumber := getCurrentBlockNumber()
			confirm := currentBlockNumber >= int64(block.Number)+confirmBlockOffset

			logger.ReloadLogger.Debugf("Transaction id: %s, blockNumber: %d, currentBlockNumber: %d, confirm: %t", txID, block.Number, currentBlockNumber, confirm)
			syncTransaction(txReceipt.TransactionHash, block, confirm)
		}
	}
}
