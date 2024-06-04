package pow

import (
	"fmt"
)

type Miner struct {
	//矿工ID
	Id int64 `json:"id"`
	//矿工账户余额
	Balance uint `json:"balance"`
	//当前矿工正在挖的区块
	blockchain *Blockchain
	// 用于通知：当接收到新区块的时候 不应该从原有的链继续往后挖
	waitForSignal chan interface{} `json:"-"`
}

// 挖矿逻辑
func (m Miner) run() {
	count := 0
	for ; ; count++ {
		//	组装区块
		blockWithoutProof := m.blockchain.assembleNewBlock(m.Id, []byte(fmt.Sprintf("模拟区块数据：%d%d", m.Id, count)))
		block, finish := blockWithoutProof.Mine(m.waitForSignal)
		if !finish {
			// 如果没有出块则继续计算
			continue
		} else {
			// 条件满足，生成新区块
			m.blockchain.AddBlock(block, m.waitForSignal)
		}
	}
}
