package pow

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	//Nonce循环上限
	maxNonce = math.MaxInt64
)

// Block 自定义区块结构
type Block struct {
	*BlockWithoutProof
	Proof
}

// Blockchain 区块链数据，因为是模拟，所以我们假设所有节点共享一条区块链数据，且所有节点共享所有矿工信息
type Blockchain struct {
	// 区块链配置信息
	config BlockchainConfig
	// 当前难度
	currentDifficulty float64
	// 区块列表
	blocks []Block
	// 矿工列表
	miners []Miner
	// 互斥锁 防止发生读写异常
	mutex *sync.RWMutex
}

// BlockchainConfig 区块链配置信息
type BlockchainConfig struct {
	MinerCount                  int     // 矿工个数
	OutBlockTime                uint    // 平均出块时间
	InitialDifficulty           float64 // 初始难度
	ModifyDifficultyBlockNumber uint    // 每多少个区块修改一次难度阈值
	BookkeepingIncentives       uint    // 记账奖励
}

// IncreaseMiner 增加矿工
func (b *Blockchain) IncreaseMiner() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	var miner = Miner{
		Id:            int64(len(b.miners)),
		Balance:       0,
		blockchain:    b,
		waitForSignal: make(chan interface{}, 1),
	}
	b.miners = append(b.miners, miner)
	go miner.run()
	return true
}

// GetBlockInfo 获取区块信息
func (b *Blockchain) GetBlockInfo() ([]Block, []Miner) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	blocks := make([]Block, len(b.blocks))
	miners := make([]Miner, len(b.miners))
	copy(blocks, b.blocks)
	copy(miners, b.miners)
	return blocks, miners
}

func (b *Blockchain) verifyNewBlock(block *Block) bool {
	prevBlock := b.blocks[len(b.blocks)-1]
	// 新区块 一定要符合 当前难度值的 要求
	if uint64(block.TargetBit) != uint64(b.currentDifficulty) {
		return false
	}
	// hash 链一定要符合
	if string(prevBlock.hash) != string(block.prevBlockHash) {
		return false
	}
	// 区块 本身需要符合规范
	if !block.Verify() {
		return false
	}
	return true
}

func (b *Block) Verify() bool {
	//

	// 检查时间戳是否合理
	if b.Proof.ActualTimestamp > time.Now().Unix() || b.Proof.ActualTimestamp < b.timestamp {
		return false
	}

	return true
}

// AddBlock 增加一个区块到区块链
func (b *Blockchain) AddBlock(block *Block, signal chan interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	block.ActualTimestamp = time.Now().Unix()
	//验证新区块
	if !b.verifyNewBlock(block) {
		return
	}

	b.blocks = append(b.blocks, *block)
	//根据挖矿难度调整难度值
	b.adjustDifficulty()
	//给予挖矿矿工奖励
	b.bookkeepingRewards(block.CoinBase)

	//通知所有矿工挖矿成功
	b.notifyMiners(block.CoinBase)

	fmt.Printf(" %s: %d 节点挖出了一个新的区块 %s\n", time.Now(), block.CoinBase, block.HashHex)
}

// 根据挖矿的时间调整难度值
func (b *Blockchain) adjustDifficulty() {
	if uint(len(b.blocks))%b.config.ModifyDifficultyBlockNumber == 0 {
		block := b.blocks[len(b.blocks)-1]
		preDiff := b.currentDifficulty
		actuallyTime := float64(block.ActualTimestamp - b.blocks[uint(len(b.blocks))-b.config.ModifyDifficultyBlockNumber].ActualTimestamp)
		theoryTime := float64(b.config.OutBlockTime * b.config.ModifyDifficultyBlockNumber)
		ratio := theoryTime / actuallyTime
		if ratio > 1.1 {
			ratio = 1.1
		} else if ratio < 0.5 {
			ratio = 0.5
		}
		b.currentDifficulty = b.currentDifficulty * ratio
		fmt.Println("难度阈值改变 preDiff: ", preDiff, "nowDiff", b.currentDifficulty)
	}
}

// 根据全局信息组装区块
func (b *Blockchain) assembleNewBlock(coinBase int64, data []byte) BlockWithoutProof {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	proof := BlockWithoutProof{
		CoinBase:         coinBase,
		timestamp:        time.Now().Unix(),
		data:             data,
		prevBlockHash:    b.blocks[len(b.blocks)-1].hash,
		TargetBit:        b.currentDifficulty,
		PrevBlockHashHex: b.blocks[len(b.blocks)-1].HashHex,
	}
	return proof
}

// 给予挖矿成功的矿工奖励
func (b *Blockchain) bookkeepingRewards(coinBase int64) {
	b.miners[coinBase].Balance += b.config.BookkeepingIncentives
}

// 通知所有矿工挖矿成功 重置矿工的Block字段
func (b *Blockchain) notifyMiners(sponsor int64) {
	for i, miner := range b.miners {
		if i != int(sponsor) {
			go func(signal chan interface{}) {
				signal <- struct{}{}
			}(miner.waitForSignal)
		}
	}
}

// RunBlockChainNetWork 运行区块链网络
func (b *Blockchain) RunBlockChainNetWork() {
	for _, m := range b.miners {
		go m.run()
	}
}

// NewBlockChainNetWork 新建区块链网络
func NewBlockChainNetWork(blockchainConfig BlockchainConfig) *Blockchain {
	b := &Blockchain{
		blocks:            nil,
		miners:            nil,
		config:            blockchainConfig,
		mutex:             &sync.RWMutex{},
		currentDifficulty: blockchainConfig.InitialDifficulty,
	}
	b.blocks = append(b.blocks, *GenerateGenesisBlock([]byte("")))
	//新建矿工
	for i := 0; i < blockchainConfig.MinerCount; i++ {
		miner := Miner{
			Id:            int64(i),
			Balance:       0,
			blockchain:    b,
			waitForSignal: make(chan interface{}, 1),
		}
		b.miners = append(b.miners, miner)
	}
	return b
}

// GenerateGenesisBlock 生成创世区块
func GenerateGenesisBlock(data []byte) *Block {
	b := &Block{BlockWithoutProof: &BlockWithoutProof{}}
	b.ActualTimestamp = time.Now().Unix()
	b.data = data
	return b
}
