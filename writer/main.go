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

	rpcInfoReader := utils.CreateChainRpcInfoReader(5 * time.Minute)
	rpcInfoReader.Start()

	// TODO: we should use here worker pool
	for _, chain := range config.SUPPORTED_CHAINS {
		time.Sleep(2 * time.Second)
		go func(chainId string) {
			for {
				rpcInfoMap, err := rpcInfoReader.GetRpcInfo()
				if err != nil {
					fmt.Printf("failed to read rpc info: %s\n", err.Error())
					time.Sleep(INTERVAL)
					continue
				}

				startTime := time.Now()

				collectBlockNumberStats(influxClient, rpcInfoMap, chainId)

				duration := time.Since(startTime)

				fmt.Printf("Collected data for chainId=%s rpcUrls=%d in %s\n", chainId, len(rpcInfoMap[chainId]), duration)

				if duration > INTERVAL {
					fmt.Printf(">>>[IMPORTANT] Collecting data take more than INTERVAL=%s duration=%s\n", INTERVAL, duration)
				}

				time.Sleep(INTERVAL)
			}
		}(chain.ChainId)
	}

	select {}
}

func collectBlockNumberStats(influxClient *influxdb3.Client, rpcsMap types.RpcInfoMap, chainId string) {
	rpcs, exists := rpcsMap[chainId]

	fmt.Printf("Collecting block number stats for chainId=%s\n", chainId)

	if !exists {
		fmt.Printf("No info in rpcMap for chainId=%s", chainId)
		return
	}

	if len(rpcs) == 0 {
		fmt.Printf("Skipping, no rpcs for chainId=%s\n", chainId)
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
