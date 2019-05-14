package nights_watch

import (
	"container/list"
	"context"
	"fmt"
	"github.com/HydroProtocol/nights-watch/ethrpc"
	"github.com/HydroProtocol/nights-watch/plugin"
	"sync"
	"time"
)

type Watcher interface {
	RegisterBlockPlugin(plugin.IBlockPlugin)
	RegisterTxPlugin(plugin.ITxPlugin)
	Run()
}

type HttpWatcher struct {
	Ctx  context.Context
	lock sync.RWMutex
	API  string
	rpc  *ethrpc.RPC

	//change to pointer element in chan
	NewBlockChan        chan ethrpc.Block
	NewTxAndReceiptChan chan *Tuple

	SyncedBlocks        *list.List
	SyncedTxAndReceipts *list.List

	BlockPlugins     []plugin.IBlockPlugin
	TxPlugins        []plugin.ITxPlugin
	TxReceiptPlugins []plugin.ITxReceiptPlugin
}

func NewHttpBasedWatcher(ctx context.Context, api string) *HttpWatcher {
	return &HttpWatcher{
		Ctx:                 ctx,
		API:                 api,
		rpc:                 ethrpc.New(api),
		NewBlockChan:        make(chan ethrpc.Block, 32),
		NewTxAndReceiptChan: make(chan *Tuple, 518),
		SyncedBlocks:        list.New(),
	}
}

func (watcher *HttpWatcher) RegisterBlockPlugin(plugin plugin.IBlockPlugin) {
	watcher.BlockPlugins = append(watcher.BlockPlugins, plugin)
}

func (watcher *HttpWatcher) RegisterTxPlugin(plugin plugin.ITxPlugin) {
	watcher.TxPlugins = append(watcher.TxPlugins, plugin)
}

func (watcher *HttpWatcher) RegisterTxReceiptPlugin(plugin plugin.ITxReceiptPlugin) {
	watcher.TxReceiptPlugins = append(watcher.TxReceiptPlugins, plugin)
}

func (watcher *HttpWatcher) LatestSyncedBlockNum() int {
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.SyncedBlocks.Len() <= 0 {
		return -1
	}

	b := watcher.SyncedBlocks.Back().Value.(ethrpc.Block)

	return b.Number
}

func (watcher *HttpWatcher) Run() {
	go func() {
		for {
			select {
			case <-watcher.Ctx.Done():
				return
			default:
				latestBlockNum, err := watcher.getCurrentBlockNum()
				fmt.Println("latestBlockNum", latestBlockNum)
				if err != nil {
					panic(err)
				}

				noNewBlockForSync := watcher.LatestSyncedBlockNum() >= latestBlockNum

				fmt.Println("watcher.LatestSyncedBlockNum()", watcher.LatestSyncedBlockNum())
				for watcher.LatestSyncedBlockNum() < latestBlockNum {
					var newBlockNumToSync int
					if watcher.LatestSyncedBlockNum() <= 0 {
						newBlockNumToSync = latestBlockNum
					} else {
						newBlockNumToSync = watcher.LatestSyncedBlockNum() + 1
					}

					newBlock, err := watcher.getBlockByNum(newBlockNumToSync)
					if err != nil {
						panic(err)
					}

					if watcher.FoundFork(newBlock) {
						fmt.Println("found fork, popping")
						watcher.popBlocksUntilReachMainChain()
					} else {
						fmt.Println("adding new block")
						watcher.addNewBlock(*newBlock)
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
	}()

	go func() {
		for {
			block := <-watcher.NewBlockChan
			// run thru block plugins
			for i := 0; i < len(watcher.BlockPlugins); i++ {
				blockPlugin := watcher.BlockPlugins[i]

				blockPlugin.AcceptBlock(block)
			}

			// run thru tx plugins
			txPlugins := watcher.TxPlugins
			for i := 0; i < len(txPlugins); i++ {
				txPlugin := txPlugins[i]

				for j := 0; j < len(block.Transactions); j++ {
					txPlugin.AcceptTx(block.Transactions[j])
				}
			}

		}
	}()

	go func() {
		for {
			tuple := <-watcher.NewTxAndReceiptChan

			txReceiptPlugins := watcher.TxReceiptPlugins
			for i := 0; i < len(txReceiptPlugins); i++ {
				txReceiptPlugin := txReceiptPlugins[i]

				txReceiptPlugin.Accept(tuple.Tx, tuple.Receipt)
			}
		}
	}()
}

func (watcher *HttpWatcher) addNewBlock(block ethrpc.Block) {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	watcher.SyncedBlocks.PushBack(block)
	watcher.NewBlockChan <- block

	// get tx receipts in block, which is time consuming
	signals := make([]*SyncSignal, 0, len(block.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		tx := block.Transactions[i]
		syncSigName := fmt.Sprintf("B:%d T:%s", block.Number, tx.Hash)

		sig := newSyncSignal(syncSigName)
		signals = append(signals, sig)

		go func() {
			txReceipt, err := watcher.rpc.EthGetTransactionReceipt(tx.Hash)
			if err != nil {
				panic(err)
			}

			sig.WaitPermission()
			watcher.NewTxAndReceiptChan <- &Tuple{&tx, txReceipt}

			sig.Done()
		}()
	}

	for i := 0; i < len(signals); i++ {
		sig := signals[i]
		sig.Permit()
		sig.WaitDone()
	}
}

type SyncSignal struct {
	name       string
	permission chan bool
	jobDone    chan bool
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

type Tuple struct {
	Tx      *ethrpc.Transaction
	Receipt *ethrpc.TransactionReceipt
}

func (watcher *HttpWatcher) popBlocksUntilReachMainChain() {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	for {
		if watcher.SyncedBlocks.Back() == nil {
			return
		}

		block, err := watcher.getBlockByNum(watcher.LatestSyncedBlockNum())
		if err != nil {
			continue
		}

		lastSyncedBlock := watcher.SyncedBlocks.Back().Value.(ethrpc.Block)

		if block.Hash != lastSyncedBlock.Hash {
			removedBlock := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Back()).(ethrpc.Block)
			removedBlock.IsRemoved = true

			tuple := watcher.SyncedTxAndReceipts.Remove(watcher.SyncedTxAndReceipts.Back()).(*Tuple)
			tuple.Tx.IsRemoved = true
			tuple.Receipt.IsRemoved = true

			watcher.NewBlockChan <- removedBlock
			watcher.NewTxAndReceiptChan <- tuple
		} else {
			return
		}
	}
}

func (watcher *HttpWatcher) FoundFork(newBlock *ethrpc.Block) bool {
	for e := watcher.SyncedBlocks.Back(); e != nil; e = e.Prev() {
		syncedBlock := e.Value.(ethrpc.Block)

		if (syncedBlock).Number+1 == newBlock.Number {
			return (syncedBlock).Hash != newBlock.ParentHash
		}
	}

	return false
}

func (watcher *HttpWatcher) getBlockByNum(blockNum int) (*ethrpc.Block, error) {
	return watcher.rpc.EthGetBlockByNumber(int(blockNum), true)
}

func (watcher *HttpWatcher) getCurrentBlockNum() (int, error) {
	i, err := watcher.rpc.EthBlockNumber()
	if err != nil {
		return -1, err
	} else {
		return i, nil
	}
}

// types and interfaces
type Hash [32]byte

type BlockWithTxInfos struct {
	ethrpc.Block
	Transactions []Transaction
}

type Transaction interface {
	Hash() Hash
}

type TransactionReceipt interface {
	Hash() Hash
	Status() bool
}
