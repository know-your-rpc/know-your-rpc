package main

import (
	"fmt"
	"koonopek/know_your_rpc/common/influx"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/writer/config"
	"koonopek/know_your_rpc/writer/stats"
	"koonopek/know_your_rpc/writer/utils"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

const (
	INTERVAL = 10 * time.Second
)

func main() {
	influxClient, err := influx.MakeInfluxDbClient()
	if err != nil {
		panic(err.Error())
	}

	defer func() {
		err := influxClient.Close()
		fmt.Printf("error while closing influx client error=%s", err.Error())
	}()

	var rpcInfoMap *types.RpcInfoMap

	utils.SetInterval(func() {
		tempRpcInfoMap, err := utils.ReadRpcInfo()
		if err != nil {
			fmt.Printf("failed to read rpc info: %s", err.Error())
		}

		rpcInfoMap = tempRpcInfoMap

		startTime := time.Now()

		for _, chain := range config.SUPPORTED_CHAINS {
			collectBlockNumberStats(influxClient, rpcInfoMap, chain.ChainId)
		}

		duration := time.Since(startTime)

		if duration > INTERVAL {
			fmt.Printf(">>>[IMPORTANT] Collecting data take more then INTERVAL=%d duration=%d \n", INTERVAL, duration)
		}
	}, INTERVAL)
}

func collectBlockNumberStats(influxClient *influxdb3.Client, rpcsMap *types.RpcInfoMap, chainId string) {
	rpcs, exists := (*rpcsMap)[chainId]

	if !exists {
		fmt.Printf("No info in rpcMap for chainId=%s", chainId)
		return
	}

	blocNumberStats := stats.BenchmarkBlockNumber(rpcs, chainId)
	bucketName := "stats-block-number"
	pointsCount, err := influx.WritePoints(influxClient, bucketName, blocNumberStats)

	if err != nil {
		fmt.Printf("failed to write points to influx\n")
	}

	fmt.Printf("wrote %d points to influx\n", pointsCount)
}
