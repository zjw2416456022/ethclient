package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	infuraURL := "https://sepolia.infura.io/v3/5b93c586b8ef48c2b4d6ee50db888e8d" // 你的 Infura 端点
	recipientAddr := "0x97afEeEF10ba9EC37f0Bdb81bF26cE6BABCdbcfE"                // 接收方地址（带0x）
	transferAmountEth := 0.001                                                   // 转账金额（ETH）

	// 测试钱包私钥地址（暂时硬编码）
	privateKeyHex := "37caacc2dfd627bbde9dd698c2c514bbcfb31c7fcea4241bc26d4f2ad60b65e7"

	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("连接 Sepolia 失败：%v", err)
	}
	defer client.Close()
	fmt.Println("成功连接到 Sepolia 测试网")

	// 解析私钥并获取发送方地址
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败：%v", err)
	}
	senderAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("发送方地址：%s\n", senderAddr.Hex())

	// 获取 Nonce（防止交易重放，账户交易序号）
	nonce, err := client.PendingNonceAt(context.Background(), senderAddr)
	if err != nil {
		log.Fatalf("获取 Nonce 失败：%v", err)
	}
	// 转换 ETH 为 Wei（1 ETH = 10^18 Wei）
	amountWei := new(big.Float).Mul(big.NewFloat(transferAmountEth), big.NewFloat(1e18))
	amountInt := new(big.Int)
	amountWei.Int(amountInt)
	// 获取推荐 Gas 价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("获取 Gas 价格失败：%v", err)
	}
	// Gas 限制（转账固定 21000）
	gasLimit := uint64(21000)
	// 解析接收方地址
	recipient := common.HexToAddress(recipientAddr)
	// 获取 Sepolia 链 ID（固定 11155111，也可动态获取）
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("获取链 ID 失败：%v", err)
	}

	// 构造未签名交易
	tx := types.NewTransaction(nonce, recipient, amountInt, gasLimit, gasPrice, nil)

	// 签名交易（EIP155 规则，绑定链 ID 防止跨链重放）
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败：%v", err)
	}

	// 发送交易到 Sepolia 测试网
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("发送交易失败：%v", err)
	}

	// 9. 输出交易哈希（可在 Sepolia Etherscan 查看）
	fmt.Println("交易发送成功！")
	fmt.Printf("交易哈希：%s\n", signedTx.Hash().Hex())
	fmt.Printf("查询地址：https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())
}
