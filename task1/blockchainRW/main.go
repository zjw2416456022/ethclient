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
	recipientAddr := "0x97afEeEF10ba9EC37f0Bdb81bF26cE6BABCdbcfE" // 接收方地址（带0x）
	transferAmountEth := 0.001                                    // 转账金额（ETH）

	// 测试环境暂时硬编码私钥
	privateKeyHex := "4eb6ac12169a4e0f836a0ffec3b01aa792fe555c67f5046eac0980d0dc488f6d" // 发送方私钥

	// Infura Sepolia 端点
	infuraURL := "https://sepolia.infura.io/v3/5b93c586b8ef48c2b4d6ee50db888e8d"

	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("连接 Sepolia 失败：%v", err)
	}
	defer client.Close()
	fmt.Println("成功连接到 Sepolia 测试网")

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败：%v（检查私钥是否为64位16进制字符串，不带0x前缀）", err)
	}
	senderAddr := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Printf("发送方地址：%s\n", senderAddr.Hex())

	if !common.IsHexAddress(recipientAddr) {
		log.Fatalf("接收方地址格式错误：%s（必须是42位16进制字符串，以0x开头）", recipientAddr)
	}

	recipient := common.HexToAddress(recipientAddr)
	fmt.Printf("接收方地址：%s（格式校验通过）\n", recipient.Hex())

	// 获取发送方已确认余额（Wei）
	balanceWei, err := client.BalanceAt(context.Background(), senderAddr, nil)
	if err != nil {
		log.Fatalf("获取发送方余额失败：%v", err)
	}
	// Wei 转换为 ETH（1 ETH = 10^18 Wei）
	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(1e18))
	fmt.Printf("💰 发送方余额：%f ETH\n", balanceEth)

	// 计算转账总费用（转账金额 + Gas费）
	// 转账金额转 Wei
	amountWei := new(big.Float).Mul(big.NewFloat(transferAmountEth), big.NewFloat(1e18))
	amountInt := new(big.Int)
	amountWei.Int(amountInt)

	// 获取推荐Gas价格（并优化：提高1.5倍，避免测试网拥堵）
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("获取 Gas 价格失败：%v", err)
	}
	// Gas价格优化：乘以3再除以2，等价于1.5倍，提高交易打包概率
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(3))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(2))
	fmt.Printf("Gas价格：%s Wei（已优化为推荐值的1.5倍）\n", gasPrice.String())

	// Gas限制（转账固定21000）
	gasLimit := uint64(21000)
	// 计算Gas费（GasLimit × GasPrice）
	gasCostWei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	gasCostEth := new(big.Float).Quo(new(big.Float).SetInt(gasCostWei), big.NewFloat(1e18))
	// 总费用 = 转账金额 + Gas费
	totalCostWei := new(big.Int).Add(amountInt, gasCostWei)
	totalCostEth := new(big.Float).Quo(new(big.Float).SetInt(totalCostWei), big.NewFloat(1e18))
	fmt.Printf("转账总费用：%f ETH（转账金额：%f ETH + Gas费：%f ETH）\n", totalCostEth, transferAmountEth, gasCostEth)

	// 检查余额是否足够
	if balanceWei.Cmp(totalCostWei) < 0 {
		log.Fatalf("余额不足！当前余额：%f ETH，需要：%f ETH", balanceEth, totalCostEth)
	}
	fmt.Println("余额校验通过，可发起转账")

	// 获取Nonce（防止交易重放）
	nonce, err := client.PendingNonceAt(context.Background(), senderAddr)
	if err != nil {
		log.Fatalf("获取 Nonce 失败：%v", err)
	}
	fmt.Printf("交易Nonce：%d\n", nonce)

	// 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("获取链 ID 失败：%v", err)
	}
	fmt.Printf("Sepolia链ID：%d\n", chainID.Uint64())

	// 构造并签名交易
	// 构造未签名交易
	tx := types.NewTransaction(nonce, recipient, amountInt, gasLimit, gasPrice, nil)
	// 签名交易（EIP155 规则，绑定链ID防止跨链重放）
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败：%v（检查私钥是否对应发送方地址）", err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("发送交易失败：%v（常见原因：余额不足/Gas过低/Nonce错误）", err)
	}

	// 输出交易结果
	fmt.Println("\n交易发送成功！")
	fmt.Printf("交易哈希：%s\n", signedTx.Hash().Hex())
	fmt.Printf("查询地址：https://sepolia.etherscan.io/tx/%s\n", signedTx.Hash().Hex())
}
