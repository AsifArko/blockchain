package main

import (
	"math/big"
	"bytes"
	"strconv"
	"fmt"
	"crypto/sha256"
	"math"
	"time"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"flag"
	"os"
)

const targetBits = 24   // dificulty of mining
const maxNonce = math.MaxInt64  // Target the upper boudary of a range . if a number (hash) is lower than the boundary its valiid , vice-versa

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type BlockChainIterator struct{
	currentHash []byte
	db *bolt.DB
}

type Blockchain struct {
	tip []byte
	db *bolt.DB
}

type ProofOfWork struct{
	block *Block
	target *big.Int
}

type CLI struct {
	bc *Blockchain
}



func NewBlock(data string , prevBlockHash []byte)*Block {
	block := &Block{time.Now().Unix(),[]byte(data),prevBlockHash,[]byte{},0}
	pow := NewProofOfWork(block)
	nonce , hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func (bc *Blockchain)AddBlock(data string) {

	blocksBucket := "Bucket"
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))
		return nil
	})

	if err!=nil{
		fmt.Println(err.Error())
	}

	newBlock := NewBlock(data ,lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash,newBlock.Serialize())
		err = b.Put([]byte("1"),newBlock.Hash)

		if err!=nil{
			fmt.Println(err.Error())
		}
		return nil
	})
}

func NewGenesisBlock()*Block{
	return NewBlock("Genesis Block",[]byte{})
}

//func NewBlockchain()*Blockchain  {
//	return &Blockchain{[]Block{*NewGenesisBlock()}}
//}

func NewBlockchain() *Blockchain{
	blocksBucket := "Bucket"
	var tip []byte
	dbFile := "BlockchainDB"
	db , err := bolt.Open(dbFile,0600,nil)

	if err != nil{
		fmt.Println(err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil{
			genesis := NewGenesisBlock()
			b,err := tx.CreateBucket([]byte(blocksBucket))
			if err!=nil{
				fmt.Println(err.Error())
			}
			err = b.Put(genesis.Hash,genesis.Serialize())
			err= b.Put([]byte("1"),genesis.Hash)
			tip = genesis.Hash
		}else {
			tip = b.Get([]byte("1"))
		}
		return nil
	})

	bc := Blockchain{tip,db}

	return &bc
}

func NewProofOfWork(b *Block) *ProofOfWork {
	// initialize big.int with value 1 and shift it left by 256 - target bits
	// 256 is the length of the SHA-256 hash in bits . if the hash is bigger than the target then its not a valid proof of work
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
		IntToHex(int64(nonce)), // nonce is the counter from the hashcash
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
	isValid := hashInt.Cmp(pow.target)==-1

	return isValid
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)

	if err != nil{
		fmt.Println(err.Error())
	}
	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&block)

	if err != nil{
		fmt.Println(err.Error())
	}
	return &block
}

func (bc *Blockchain) Iterator() *BlockChainIterator{
	bc1 := &BlockChainIterator{bc.tip,bc.db}
	return bc1
}

func (i *BlockChainIterator)Next()*Block {
	var block *Block
	blocksBucket := "Bucket"
	err := i.db.View(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if err != nil{
		fmt.Println(err.Error())
	}

	i.currentHash = block.PrevBlockHash

	return block
}

func (c *CLI)printUsage()  {
	fmt.Println("go run filename transaction")
}

func checkError(err error)  {
	if err!=nil{
		fmt.Println(err.Error())
	}
}

func (c *CLI)addBlock(data string)  {
	c.bc.AddBlock(data)
	fmt.Println("Success")
}

func (c *CLI)printchain() {
	bci := c.bc.Iterator()

	for{
		block := bci.Next()
		fmt.Printf("Prev. hash : %x\n",block.PrevBlockHash)
		fmt.Printf("Data : %s\n",block.Data)
		fmt.Printf("Hash : %x\n",block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("Pow : %s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0{
			break
		}

	}
}

func (cli *CLI) Run() {
	addBlockCmd   := flag.NewFlagSet("addblock",flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain",flag.ExitOnError)
	addBlockData  := addBlockCmd.String("data","","Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		checkError(err)
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		checkError(err)
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed(){
		if *addBlockData == ""{
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed(){
		cli.printchain()
	}
}

func main()  {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Evan")
	bc.AddBlock("Send 2 more BTC to Arko")
	bc.AddBlock("Send 2 more BTC to Harami")

	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()

}