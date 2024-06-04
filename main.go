package main

import (
	"POWdemo/pow"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type BlockchainInfo struct {
	Blocks []*pow.Block `json:"blocks"` // 区块列表
	Miners []*pow.Miner `json:"miners"` // 矿工列表
}

func main() {
	var count int
	fmt.Printf("请输入初始矿工数量：")
	fmt.Scanf("%d", &count)
	time.Sleep(10000)
	fmt.Printf("开始挖矿")
	//新建区块链网络
	work := pow.NewBlockChainNetWork(pow.BlockchainConfig{
		//矿工数量
		MinerCount: count,
		//平均出块时间
		OutBlockTime: 10,
		//初始难道值
		InitialDifficulty: 20,
		//每多少个区块修改一次难度值
		ModifyDifficultyBlockNumber: 10,
		//每次记账奖励
		BookkeepingIncentives: 20,
	})
	//运行区块链网络
	work.RunBlockChainNetWork()
	//启动web服务
	RunRouter(work)

}

// 增加矿工
func addMiner(blockchain *pow.Blockchain) gin.HandlerFunc {
	return func(c *gin.Context) {
		blockchain.IncreaseMiner()
		c.JSON(200, gin.H{
			"message": "增加成功",
		})
	}
}

// 打印挖矿信息
func getBlockChainInfo(blockchain *pow.Blockchain) gin.HandlerFunc {
	return func(c *gin.Context) {
		blocks, miners := blockchain.GetBlockInfo()
		c.JSON(200, gin.H{
			"blocks": blocks,
			"miners": miners,
		})
	}
}

// RunRouter 运行web服务
// 访问localhost:8080/addMiner 可以增加矿工
// 访问localhost:8080/getBlockChainInfo获取到目前为止打印的挖矿信息
func RunRouter(blockchain *pow.Blockchain) {
	r := gin.Default()
	r.GET("/addMiner", addMiner(blockchain))
	r.GET("/getBlockChainInfo", getBlockChainInfo(blockchain))
	r.Run()
}
