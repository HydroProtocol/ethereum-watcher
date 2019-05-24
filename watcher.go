package nights_watch

import (
	"container/list"
	"context"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/ethereum"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/common/log"
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
//	Run()
//}

type IBlockChainRPC interface {
	GetCurrentBlockNum() (uint64, error)

	GetBlockByNum(uint64) (sdk.Block, error)
	GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error)
}

type EthBlockChainRPC struct {
	rpcImpl *ethrpc.EthRPC
}

func (rpc EthBlockChainRPC) GetBlockByNum(num uint64) (sdk.Block, error) {
	b, err := rpc.rpcImpl.EthGetBlockByNumber(int(num), true)
	return &ethereum.EthereumBlock{b}, err
	panic("implement me")
}

func (rpc EthBlockChainRPC) GetTransactionReceipt(txHash string) (sdk.TransactionReceipt, error) {
	receipt, error := rpc.rpcImpl.EthGetTransactionReceipt(txHash)
	return &ethereum.EthereumTransactionReceipt{receipt}, error
}

func (rpc EthBlockChainRPC) GetCurrentBlockNum() (uint64, error) {
	num, err := rpc.rpcImpl.EthBlockNumber()
	return uint64(num), err
}

func NewEthRPC(api string) *EthBlockChainRPC {
	rpc := ethrpc.New(api)

	return &EthBlockChainRPC{rpc}
}

type AbstractWatcher struct {
	//IWatcher

	//API  string
	//rpc  *ethrpc.EthRPC
	rpc IBlockChainRPC

	Ctx  context.Context
	lock sync.RWMutex

	//change to pointer element in chan
	NewBlockChan        chan *structs.RemovableBlock
	NewTxAndReceiptChan chan *structs.RemovableTxAndReceipt

	SyncedBlocks        *list.List
	SyncedTxAndReceipts *list.List

	BlockPlugins     []plugin.IBlockPlugin
	TxPlugins        []plugin.ITxPlugin
	TxReceiptPlugins []plugin.ITxReceiptPlugin
}

func NewHttpBasedEthWatcher(ctx context.Context, api string) *AbstractWatcher {
	rpc := NewEthRPC(api)

	return &AbstractWatcher{
		Ctx:                 ctx,
		rpc:                 rpc,
		NewBlockChan:        make(chan *structs.RemovableBlock, 32),
		NewTxAndReceiptChan: make(chan *structs.RemovableTxAndReceipt, 518),
		SyncedBlocks:        list.New(),
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

func (watcher *AbstractWatcher) LatestSyncedBlockNum() uint64 {
	watcher.lock.RLock()
	defer watcher.lock.RUnlock()

	if watcher.SyncedBlocks.Len() <= 0 {
		return 0
	}

	b := watcher.SyncedBlocks.Back().Value.(sdk.Block)

	return b.Number()
}

func (watcher *AbstractWatcher) Run() error {
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
		for tuple := range watcher.NewTxAndReceiptChan {
			txReceiptPlugins := watcher.TxReceiptPlugins
			for i := 0; i < len(txReceiptPlugins); i++ {
				txReceiptPlugin := txReceiptPlugins[i]

				txReceiptPlugin.Accept(structs.NewRemovableTxAndReceipt(tuple.Tx, tuple.Receipt, false))
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
			maxRetryCount := 3

			var latestBlockNum uint64
			var err error

			for i := 0; i < maxRetryCount; i++ {
				latestBlockNum, err = watcher.rpc.GetCurrentBlockNum()
				if err == nil {
					break
				}
			}
			if err != nil {
				return err
			}

			fmt.Println("latestBlockNum", latestBlockNum)

			noNewBlockForSync := watcher.LatestSyncedBlockNum() >= latestBlockNum

			fmt.Println("watcher.LatestSyncedBlockNum()", watcher.LatestSyncedBlockNum())
			for watcher.LatestSyncedBlockNum() < latestBlockNum {
				var newBlockNumToSync uint64
				if watcher.LatestSyncedBlockNum() <= 0 {
					newBlockNumToSync = latestBlockNum
				} else {
					newBlockNumToSync = watcher.LatestSyncedBlockNum() + 1
				}

				newBlock, err := watcher.rpc.GetBlockByNum(newBlockNumToSync)
				if err != nil {
					panic(err)
				}

				if watcher.FoundFork(newBlock) {
					fmt.Println("found fork, popping")
					watcher.popBlocksUntilReachMainChain()
				} else {
					fmt.Println("adding new block")
					watcher.addNewBlock(structs.NewRemovableBlock(newBlock, false))
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

func (watcher *AbstractWatcher) addNewBlock(block *structs.RemovableBlock) {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	watcher.SyncedBlocks.PushBack(block)
	watcher.NewBlockChan <- block

	// get tx receipts in block, which is time consuming
	signals := make([]*SyncSignal, 0, len(block.GetTransactions()))
	for i := 0; i < len(block.GetTransactions()); i++ {
		tx := block.GetTransactions()[i]
		syncSigName := fmt.Sprintf("B:%d T:%s", block.Number(), tx.GetHash())

		sig := newSyncSignal(syncSigName)
		signals = append(signals, sig)

		go func() {
			var txReceipt sdk.TransactionReceipt
			var err error

			retry := 3
			for i := 0; i <= retry; i++ {
				txReceipt, err = watcher.rpc.GetTransactionReceipt(tx.GetHash())
				if err == nil {
					break
				}
			}

			if err != nil {
				panic(fmt.Sprintf("GetTransactionReceipt fail after %d-times retry, err: %s", retry, err))
			}

			sig.WaitPermission()

			watcher.NewTxAndReceiptChan <- structs.NewRemovableTxAndReceipt(tx, txReceipt, false)

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

func (watcher *AbstractWatcher) popBlocksUntilReachMainChain() {
	watcher.lock.Lock()
	defer watcher.lock.Unlock()

	for {
		log.Debug("check tail block:", watcher.SyncedBlocks.Back())
		if watcher.SyncedBlocks.Back() == nil {
			return
		}

		block, err := watcher.rpc.GetBlockByNum(watcher.LatestSyncedBlockNum())
		if err != nil {
			continue
		}

		lastSyncedBlock := watcher.SyncedBlocks.Back().Value.(structs.RemovableBlock)

		if block.Hash() != lastSyncedBlock.Hash() {
			log.Debug("removing tail block:", watcher.SyncedBlocks.Back())
			removedBlock := watcher.SyncedBlocks.Remove(watcher.SyncedBlocks.Back()).(sdk.Block)

			log.Debug("removing tail txAndReceipt:", watcher.SyncedTxAndReceipts.Back())
			tuple := watcher.SyncedTxAndReceipts.Remove(watcher.SyncedTxAndReceipts.Back()).(*structs.TxAndReceipt)

			watcher.NewBlockChan <- structs.NewRemovableBlock(removedBlock, true)
			watcher.NewTxAndReceiptChan <- structs.NewRemovableTxAndReceipt(tuple.Tx, tuple.Receipt, true)
		} else {
			return
		}
	}
}

func (watcher *AbstractWatcher) FoundFork(newBlock sdk.Block) bool {
	for e := watcher.SyncedBlocks.Back(); e != nil; e = e.Prev() {
		syncedBlock := e.Value.(sdk.Block)

		if (syncedBlock).Number()+1 == newBlock.Number() {
			notMatch := (syncedBlock).Hash() != newBlock.ParentHash()

			if notMatch {
				log.Debugf("found fork, new block(%d): %s, new block's parent: %s, parent we synced: %s",
					newBlock.Number(), newBlock.Hash(), newBlock.ParentHash(), syncedBlock.Hash())

				return true
			}
		}
	}

	return false
}
