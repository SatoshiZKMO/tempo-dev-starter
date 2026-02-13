package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	blockStateFile     = "last_block.txt"
	analyticsStateFile = "analytics_state.json"
)

type Token struct {
	Name     string
	Address  common.Address
	Decimals int
}

type LifetimeStats struct {
	Transfers int     `json:"transfers"`
	Incoming  float64 `json:"incoming"`
	Outgoing  float64 `json:"outgoing"`
	Fees      float64 `json:"fees"`
}

func main() {

	rpcURL := "https://rpc.moderato.tempo.xyz"
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}

	// 👇 BURAYA WALLET ADRESİNİ KOY
	targetAddress := common.HexToAddress("0xAE14787a6b607A118A1f3185fFe38fb5451f7768")

	tokens := []Token{
		{"pathUSD", common.HexToAddress("0x20c0000000000000000000000000000000000000"), 6},
		{"AlphaUSD", common.HexToAddress("0x20c0000000000000000000000000000000000001"), 6},
		{"BetaUSD", common.HexToAddress("0x20c0000000000000000000000000000000000002"), 6},
		{"ThetaUSD", common.HexToAddress("0x20c0000000000000000000000000000000000003"), 6},
	}

	fmt.Println("Tempo Persistent Multi-Token Analytics Engine Started...")

	for {
		latestBlock, _ := client.BlockNumber(context.Background())
		lastProcessed := loadLastBlock(latestBlock)

		if latestBlock > lastProcessed {

			fmt.Println("New blocks:", lastProcessed, "->", latestBlock)

			stats := loadAnalytics()

			for _, token := range tokens {
				processToken(client, token, targetAddress, lastProcessed, latestBlock, stats)
			}

			printGlobalMetrics(stats)

			saveAnalytics(stats)
			saveLastBlock(latestBlock)
		}

		time.Sleep(5 * time.Second)
	}
}

func processToken(client *ethclient.Client, token Token, target common.Address, from uint64, to uint64, stats map[string]LifetimeStats) {

	transferEventSig := []byte("Transfer(address,address,uint256)")
	hash := crypto.Keccak256Hash(transferEventSig)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{token.Address},
		Topics:    [][]common.Hash{{hash}},
	}

	logs, _ := client.FilterLogs(context.Background(), query)

	erc20ABI := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
	parsedABI, _ := abi.JSON(strings.NewReader(erc20ABI))

	feeAddress := common.HexToAddress("0xfeEC000000000000000000000000000000000000")

	current := stats[token.Name]

	for _, vLog := range logs {

		fromAddr := common.HexToAddress(vLog.Topics[1].Hex())
		toAddr := common.HexToAddress(vLog.Topics[2].Hex())

		var event struct {
			Value *big.Int
		}
		parsedABI.UnpackIntoInterface(&event, "Transfer", vLog.Data)

		valueFloat := toHuman(event.Value, token.Decimals)

		// Fee
		if toAddr == feeAddress && fromAddr == target {
			current.Fees += valueFloat
			continue
		}

		// Outgoing
		if fromAddr == target && toAddr != feeAddress {
			current.Outgoing += valueFloat
			current.Transfers++
		}

		// Incoming
		if toAddr == target && fromAddr != target {
			current.Incoming += valueFloat
			current.Transfers++
		}
	}

	stats[token.Name] = current

	if current.Transfers > 0 {
		fmt.Println("==== Lifetime Stats:", token.Name, "====")
		fmt.Println("Transfers:", current.Transfers)
		fmt.Println("Incoming:", current.Incoming)
		fmt.Println("Outgoing:", current.Outgoing)
		fmt.Println("Fees Paid:", current.Fees)
		fmt.Println("================================")
	}
}

func printGlobalMetrics(stats map[string]LifetimeStats) {

	var totalTransfers int
	var totalVolume float64
	var totalFees float64

	for _, s := range stats {
		totalTransfers += s.Transfers
		totalVolume += s.Outgoing
		totalFees += s.Fees
	}

	var feeRate float64
	if totalVolume > 0 {
		feeRate = (totalFees / totalVolume) * 100
	}

	fmt.Println("===== GLOBAL METRICS =====")
	fmt.Println("Total Transfers:", totalTransfers)
	fmt.Println("Total Volume:", totalVolume)
	fmt.Println("Total Fees:", totalFees)
	fmt.Printf("Effective Fee Rate: %.4f%%\n", feeRate)
	fmt.Println("==========================")
}

func toHuman(value *big.Int, decimals int) float64 {
	divisor := float64(pow10(decimals))
	valFloat, _ := new(big.Float).Quo(
		new(big.Float).SetInt(value),
		new(big.Float).SetFloat64(divisor),
	).Float64()
	return valFloat
}

func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

func loadLastBlock(latest uint64) uint64 {
	data, err := os.ReadFile(blockStateFile)
	if err != nil {
		return latest - 5
	}
	block, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return latest - 5
	}
	return block
}

func saveLastBlock(block uint64) {
	os.WriteFile(blockStateFile, []byte(strconv.FormatUint(block, 10)), 0644)
}

func loadAnalytics() map[string]LifetimeStats {
	stats := make(map[string]LifetimeStats)
	data, err := os.ReadFile(analyticsStateFile)
	if err != nil {
		return stats
	}
	json.Unmarshal(data, &stats)
	return stats
}

func saveAnalytics(stats map[string]LifetimeStats) {
	data, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(analyticsStateFile, data, 0644)
}
