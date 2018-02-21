package main

import (
	"strconv"
	"bytes"
	"crypto/sha256"

	"time"
	"fmt"
)

type Blockchain struct {
	blocks []Block
}

type Block struct {
	Timestamp int64 `Current Time stamp`
	Data []byte `Actual Data`
	PrevBlockHash []byte `Previous block Hash`
	Hash []byte `Hash of the block`
}

func (b *Block)SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp,10))
	headers := bytes.Join([][]byte{b.PrevBlockHash,b.Data,timestamp},[]byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

func NewBlock(data string , PrevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(),[]byte(data),PrevBlockHash,[]byte{},}
	block.SetHash()
	return block
}

func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, *newBlock)
}

func NewGenesisBlock()*Block{
	return NewBlock("Genesis Block",[]byte{})
}

func NewBlockchain()*Blockchain  {
	return &Blockchain{[]Block{*NewGenesisBlock()}}
}

func main()  {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Evan")
	bc.AddBlock("Send 2 more BTC to Evan")

	for _,block := range bc.blocks{
		fmt.Println("Timestamp : ",block.Timestamp)
		fmt.Printf("Prev Hash : %x\n",block.PrevBlockHash)
		fmt.Printf("Data : %s\n",block.Data)
		fmt.Printf("Hash : %x\n",block.Hash)
		fmt.Println()
	}

}