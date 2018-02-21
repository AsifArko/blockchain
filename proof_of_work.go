package main

import (
	"math/big"
	"bytes"
	"strconv"
	"fmt"
	"crypto/sha256"
	"math"
	"time"
)

const targetBits = 24
const maxNonce = math.MaxInt64

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type Blockchain struct {
	blocks []Block
}

type ProofOfWork struct{
	block *Block
	target *big.Int
}

func NewBlock(data string , prevBlockHash []byte)*Block {
	block := &Block{time.Now().Unix(),[]byte(data),prevBlockHash,[]byte{},0}
	pow := NewProofOfWork(block)
	nonce , hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

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

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))
	pow := &ProofOfWork{b,target}

	return pow
}

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func (pow *ProofOfWork)prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		IntToHex(pow.block.Timestamp),
		IntToHex(int64(targetBits)),
		IntToHex(int64(nonce)),

	},[]byte{})

	return data
}

func (pow *ProofOfWork)Run()(int,[]byte)  {
	start := time.Now()

	var hashInt big.Int
	var hash [32]byte
	nonce :=0

	fmt.Printf("Mining the block containing \"%s\"\n",pow.block.Data)

	for nonce < maxNonce{
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target)==-1{
			break
		}else{
			nonce++
		}
	}
	fmt.Print("  ",time.Now().Sub(start),"\n\n")
	return nonce,hash[:]
}

func (pow *ProofOfWork)Validate()bool  {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target)==1

	return isValid
}

func main()  {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Evan")
	bc.AddBlock("Send 2 more BTC to Arko")
	bc.AddBlock("Send 5 BTC to Jhon")
	bc.AddBlock("Send 7 more BTC to Sap")
	bc.AddBlock("Send 8 BTC to Masum")
	bc.AddBlock("Send 9 more BTC to Fahim")
	bc.AddBlock("Send 1 BTC to Rejve")
	bc.AddBlock("Send 2 more BTC to Harami")

	for _,block := range bc.blocks{
		pow := NewProofOfWork(&block)
		fmt.Println("Timestamp : ",block.Timestamp)
		fmt.Printf("Prev Hash : %x\n",block.PrevBlockHash)
		fmt.Printf("Data : %s\n",block.Data)
		fmt.Printf("Hash : %x\n",block.Hash)
		fmt.Printf("Pow : %s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}

}