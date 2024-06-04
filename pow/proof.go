package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

// Proof 区块的证明信息
type Proof struct {
	//实际的时间戳 由于比特币在挖矿中不光要变动nonce值 也要变动时间戳
	ActualTimestamp int64 `json:"actualTimestamp"`
	//随机值
	Nonce int64 `json:"nonce"`
	//当前块哈希
	hash []byte
	// 转换成十六进制可读
	HashHex string `json:"hashHex"`
}

// BlockWithoutProof 不带证明信息的区块
type BlockWithoutProof struct {
	// 挖矿成功矿工
	CoinBase int64 `json:"coinBase"`
	//时间戳
	timestamp int64
	//数据域
	data []byte
	//前一块hash
	prevBlockHash []byte
	//前一块hash
	PrevBlockHashHex string `json:"prevBlockHashHex"`
	//目标阈值
	TargetBit float64 `json:"targetBit"`
}

// 准备数据 整理成待计算哈希
func (bf *BlockWithoutProof) prepareData(nonce int64) []byte {
	data := bytes.Join(
		[][]byte{
			int2Hex(bf.CoinBase),
			bf.prevBlockHash,
			bf.data,
			int2Hex(bf.timestamp),
			int2Hex(int64(bf.TargetBit)),
			int2Hex(nonce),
		},
		[]byte{},
	)

	return data
}

// Mine 挖矿函数
func (bf *BlockWithoutProof) Mine(waitForSignal chan interface{}) (*Block, bool) {
	//target为最终难度值
	target := big.NewInt(1)
	//target为1向左位移256-24（挖矿难度）
	target.Lsh(target, uint(256-bf.TargetBit))

	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	for nonce != maxNonce {
		// 判断一下是否别的矿工已经计算出来结果了 模拟 一旦收到其他矿工		的交易，立即停止计算
		select {
		case _ = <-waitForSignal:
			return nil, false
		default:
			//准备数据整理为哈希
			data := bf.prepareData(int64(nonce))
			//计算哈希
			hash = sha256.Sum256(data)
			hashInt.SetBytes(hash[:])
			//按字节比较，hashInt cmp小于0代表找到目标Nonce
			if hashInt.Cmp(target) < 0 {
				block := &Block{
					BlockWithoutProof: bf,
					Proof: Proof{
						Nonce:   int64(nonce),
						hash:    hash[:],
						HashHex: hex.EncodeToString(hash[:]),
					},
				}
				return block, true
			} else {
				nonce++
			}
		}
	}
	return nil, false
}

// int2Hex 将 int64 转换为字节数组
func int2Hex(num int64) []byte {
	buff := make([]byte, 8)                       // 创建一个长度为8的字节数组，因为 int64 占8个字节
	binary.BigEndian.PutUint64(buff, uint64(num)) // 使用大端序将 num 写入字节数组
	return buff
}
