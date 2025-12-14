package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	counter "github.com/zjw2416456022/ethclient"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("加载.env文件失败：%v", err)
	}
	rpcURL := os.Getenv("SEPOLIA_RPC_URL")
	privateKeyStr := os.Getenv("PRIVATE_KEY")
	contractAddrStr := os.Getenv("CONTRACT_ADDRESS")

	// 校验配置是否为空
	if rpcURL == "" || privateKeyStr == "" || contractAddrStr == "" {
		log.Fatal("请检查.env文件，确保SEPOLIA_RPC_URL、PRIVATE_KEY、CONTRACT_ADDRESS都已配置")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接Sepolia失败：%v", err)
	}
	defer client.Close()

	// 加载钱包私钥
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatalf("解析私钥失败：%v", err)
	}

	// 配置交易选项（gas价格、链ID等）
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("获取链ID失败：%v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("创建交易器失败：%v", err)
	}
	// 设置Gas上限（可根据网络调整）
	auth.GasLimit = uint64(300000)

	// 连接已部署的合约
	contractAddr := common.HexToAddress(contractAddrStr)
	counterContract, err := counter.NewCounter(contractAddr, client)
	if err != nil {
		log.Fatalf("连接合约失败：%v", err)
	}

	// 调用合约方法：先获取当前计数
	initialCount, err := counterContract.GetCount(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Fatalf("获取计数失败：%v", err)
	}
	fmt.Printf("调用前计数器值：%d\n", initialCount)

	// 调用合约方法：增加计数（+1）
	tx, err := counterContract.Inc(auth)
	if err != nil {
		log.Fatalf("调用Increment失败：%v", err)
	}
	fmt.Printf("交易已发送，哈希：%s\n", tx.Hash().Hex())

	// 等待交易确认，再获取新计数
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("等待交易确认失败：%v", err)
	}

	newCount, err := counterContract.GetCount(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		log.Fatalf("获取新计数失败：%v", err)
	}
	fmt.Printf("调用后计数器值：%d\n", newCount)
}
