package nights_watch

import (
	"container/list"
	"context"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/rpc"
	"github.com/HydroProtocol/nights-watch/structs"
	"sync"
	"time"
)

//type IWatcher interface {
//	RegisterBlockPlugin(plugin.IBlockPlugin)
//	RegisterTxPlugin(plugin.ITxPlugin)
//	RegisterTxReceiptPlugin(plugin.ITxReceiptPlugin)
//
//	//GetCurrentBlockNum() (uint64, error)
//	//GetCurrentBlock() (sdk.Block, error)
//	//GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error)
//
//	RunTillExit()
//}

type AbstractWatcher struct {
	//IWatcher

	rpc rpc.IBlockChainRPC

	Ctx  context.Context
	lock sync.RWMutex

	NewBlockChan        chan *structs.RemovableBlock
	NewTxAndReceiptChan chan *structs.RemovableTxAndReceipt

	SyncedBlocks        *list.List
	SyncedTxAndReceipts *list.List

	BlockPlugins     []plugin.IBlockPlugin
	TxPlugins        []plugin.ITxPlugin
	TxReceiptPlugins []plugin.ITxReceiptPlugin
}

func NewHttpBasedEthWatcher(ctx context.Context, api string) *AbstractWatcher {
	rpc := rpc.NewEthRPCWithRetry(api, 5)

	return &AbstractWatcher{
		Ctx:                 ctx,
		rpc:                 rpc,
		NewBlockChan:        make(chan *structs.RemovableBlock, 32),
		NewTxAndReceiptChan: make(chan *structs.RemovableTxAndReceipt, 518),
		SyncedBlocks:        list.New(),
		SyncedTxAndReceipts: list.New(),
	}
}

func (watcher *AbstractWatcher) RegisterBlockPlugin(plugin plugin.IBlockPlugin) {
	watcher.BlockPlugins = append(watcher.BlockPlugins, plugin)
}

func (watcher *AbstractWatcher) RegisterTxPlugin(plugin plugin.ITxPlugin) {
	watcher.TxPlugins = append(watcher.TxPlugins, plugin)
}

func (watcher *AbstractWatcher) RegisterTxReceiptPlugin(plugin plugin.ITxReceiptPlugin) {
	watcher.TxReceiptPlugins = append(watcher.TxReceiptPlugins, plugin)
}

// start sync from latest block
func (watcher *AbstractWatcher) RunTillExit() error {
	return watcher.RunTillExitFromBlock(0)
}

// start sync from given block
// 0 means start from latest block
func (watcher *AbstractWatcher) RunTillExitFromBlock(startBlockNum uint64) error {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		for block := range watcher.NewBlockChan {
			// run thru block plugins
			for i := 0; i < len(watcher.BlockPlugins); i++ {
				blockPlugin := watcher.BlockPlugins[i]

				blockPlugin.AcceptBlock(block)
			}

			// run thru tx plugins
			txPlugins := watcher.TxPlugins
			for i := 0; i < len(txPlugins); i++ {
				txPlugin := txPlugins[i]

				for j := 0; j < len(block.GetTransactions()); j++ {
					tx := structs.NewRemovableTx(block.GetTransactions()[j], false)
					txPlugin.AcceptTx(tx)
				}
			}
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for removableTxAndReceipt := range watcher.NewTxAndReceiptChan {

			txReceiptPlugins := watcher.TxReceiptPlugins
			for i := 0; i < len(txReceiptPlugins); i++ {
				txReceiptPlugin := txReceiptPlugins[i]

				if p, ok := txReceiptPlugin.(*plugin.TxReceiptPluginWithFilter); ok {
					// for filter plugin, only feed receipt it wants
					if p.NeedReceipt(removableTxAndReceipt.Tx) {
						txReceiptPlugin.Accept(removableTxAndReceipt)
					}
				} else {
					txReceiptPlugin.Accept(removableTxAndReceipt)
				}
			}
		}

		wg.Done()
	}()

	for {
		select {
		case <-watcher.Ctx.Done():
			close(watcher.NewBlockChan)
			close(watcher.NewTxAndReceiptChan)

			wg.Wait()

			return nil
		default:
			latestBlockNum, err := watcher.rpc.GetCurrentBlockNum()
			if err != nil {
				return err
			}

			if startBlockNum <= 0 {
				startBlockNum = latestBlockNum
			}

			noNewBlockForSync := watcher.LatestSyncedBlockNum() >= latestBlockNum

			fmt.Println("watcher.LatestSyncedBlockNum()", watcher.LatestSyncedBlockNum())
			for watcher.LatestSyncedBlockNum() < latestBlockNum {
				var newBlockNumToSync uint64
				if watcher.LatestSyncedBlockNum() <= 0 {
					newBlockNumToSync = startBlockNum
				} else {
					newBlockNumToSync = watcher.LatestSyncedBlockNum() + 1
				}

				newBlock, err := watcher.rpc.GetBlockByNum(newBlockNumToSync)
				if err != nil {
					return err
				}

				if watcher.FoundFork(newBlock) {
					//todo seems blocked, bug?
					fmt.Println("found fork, popping")
					err = watcher.popBlocksUntilReachMainChain()
				} else {
					fmt.Println("adding new block")
					err = watcher.addNewBlock(structs.NewRemovableBlock(newBlock, false))
				}

				if err != nil {
					return err
				}
			}

			if noNewBlockForSync {
				fmt.Println("no new block to sync, sleep for 3 secs")

				// sleep for 3 secs
				timer := time.NewTimer(3 * time.Second)
				<-timer.C
			}
		}
	}
}

func (watcher *AbstractWatcher) LatestSyncedBlockNum() uint64 {
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.SyncedBlocks.Len() <= 0 {
		return 0
	}

	b := watcher.SyncedBlocks.Back().Value.(sdk.Block)

	return b.Number()
}

// go thru plugins to check if this watcher need fetch receipt for tx
// network load for fetching receipts per tx is heavy,
// we use this method to make sure we only do the work we need
func (watcher *AbstractWatcher) needReceipt(tx sdk.Transaction) bool {
	plugins := watcher.TxReceiptPlugins

	for _, p := range plugins {
		if filterPlugin, ok := p.(plugin.TxReceiptPluginWithFilter); ok {
			if filterPlugin.NeedReceipt(tx) {
				return true
			}
		} else {
			// exist global tx-receipt listener
			return true
		}
	}

	return false
}

func (watcher *AbstractWatcher) addNewBlock(block *structs.RemovableBlock) error {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	// get tx receipts in block, which is time consuming
	signals := make([]*SyncSignal, 0, len(block.GetTransactions()))
	for i := 0; i < len(block.GetTransactions()); i++ {
		tx := block.GetTransactions()[i]

		if !watcher.needReceipt(tx) {
			fmt.Printf("ignored tx: %s", tx.GetHash())
			continue
		}

		syncSigName := fmt.Sprintf("B:%d T:%s", block.Number(), tx.GetHash())

		sig := newSyncSignal(syncSigName)
		signals = append(signals, sig)

		go func() {
			txReceipt, err := watcher.rpc.GetTransactionReceipt(tx.GetHash())

			if err != nil {
				fmt.Printf("GetTransactionReceipt fail, err: %s", err)
				sig.err = err

				// one fails all
				return
			}

			sig.WaitPermission()

			sig.rst = structs.NewRemovableTxAndReceipt(tx, txReceipt, false)

			sig.Done()
		}()
	}

	for i := 0; i < len(signals); i++ {
		sig := signals[i]
		sig.Permit()
		sig.WaitDone()

		if sig.err != nil {
			return sig.err
		}
	}

	for i := 0; i < len(signals); i++ {
		watcher.SyncedTxAndReceipts.PushBack(signals[i].rst.TxAndReceipt)
		watcher.NewTxAndReceiptChan <- signals[i].rst
	}

	// block
	watcher.SyncedBlocks.PushBack(block.Block)
	watcher.NewBlockChan <- block

	return nil
}

type SyncSignal struct {
	name       string
	permission chan bool
	jobDone    chan bool
	rst        *structs.RemovableTxAndReceipt
	err        error
}

func newSyncSignal(name string) *SyncSignal {
	return &SyncSignal{
		name:       name,
		permission: make(chan bool, 1),
		jobDone:    make(chan bool, 1),
	}
}

func (s *SyncSignal) Permit() {
	s.permission <- true
}

func (s *SyncSignal) WaitPermission() {
	<-s.permission
}

func (s *SyncSignal) Done() {
	s.jobDone <- true
}

func (s *SyncSignal) WaitDone() {
	<-s.jobDone
}

func (watcher *AbstractWatcher) popBlocksUntilReachMainChain() error {
	fmt.Println("before lock popBlocksUntilReachMainChain")
	watcher.lock.Lock()
	defer watcher.lock.Unlock()
	fmt.Println("after lock popBlocksUntilReachMainChain")

	for {
		fmt.Println("check tail block:", watcher.SyncedBlocks.Back())
		if watcher.SyncedBlocks.Back() == nil {
			return nil
		}

		fmt.Println("1")
		// NOTE: instead of watcher.LatestSyncedBlockNum() cuz it has lock
		lastSyncedBlock := watcher.SyncedBlocks.Back().Value.(sdk.Block)
		block, err := watcher.rpc.GetBlockByNum(lastSyncedBlock.Number())
		if err != nil {
			return err
		}

		fmt.Println("3")
		if block.Hash() != lastSyncedBlock.Hash() {
			fmt.Println("4")
			fmt.Println("removing tail block:", watcher.SyncedBlocks.Back())
			removedBlock := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Back()).(sdk.Block)

			for watcher.SyncedTxAndReceipts.Back() != nil {

				tail := watcher.SyncedTxAndReceipts.Back()

				if tail.Value.(*structs.TxAndReceipt).Tx.GetBlockNumber() >= removedBlock.Number() {
					fmt.Println("removing tail txAndReceipt:", tail)
					tuple := watcher.SyncedTxAndReceipts.Remove(tail).(*structs.TxAndReceipt)

					fmt.Println("6")
					watcher.NewTxAndReceiptChan <- structs.NewRemovableTxAndReceipt(tuple.Tx, tuple.Receipt, true)
					fmt.Println("7")
				} else {
					break
				}
			}

			fmt.Println("5")
			watcher.NewBlockChan <- structs.NewRemovableBlock(removedBlock, true)
		} else {
			fmt.Println("44")
			return nil
		}
	}
}

func (watcher *AbstractWatcher) FoundFork(newBlock sdk.Block) bool {
	for e := watcher.SyncedBlocks.Back(); e != nil; e = e.Prev() {
		syncedBlock := e.Value.(sdk.Block)

		if (syncedBlock).Number()+1 == newBlock.Number() {
			notMatch := (syncedBlock).Hash() != newBlock.ParentHash()

			if notMatch {
				fmt.Printf("found fork, new block(%d): %s, new block's parent: %s, parent we synced: %s",
					newBlock.Number(), newBlock.Hash(), newBlock.ParentHash(), syncedBlock.Hash())

				return true
			}
		}
	}

	return false
}
