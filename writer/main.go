package main

import (
	"fmt"
	"koonopek/know_your_rpc/writer/influx"
	"koonopek/know_your_rpc/writer/stats"
	"koonopek/know_your_rpc/writer/utils"
	"os"
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

	rpcInfoMap, err := utils.ReadRpcInfo()

	if err != nil {
		panic(err.Error())
	}

	// collect data for eth every 1min
	// todo could by optimize by writing points only once
	setInterval(func() {
		collectBlockNumberStats(influxClient, rpcInfoMap, "1")
		collectBlockNumberStats(influxClient, rpcInfoMap, "56")
		collectBlockNumberStats(influxClient, rpcInfoMap, "42161")
	}, INTERVAL)
}

func collectBlockNumberStats(influxClient *influxdb3.Client, rpcsMap *utils.RpcInfoMap, chainId string) {
	rpcs, exists := (*rpcsMap)[chainId]

	if !exists {
		fmt.Printf("No info in rpcMap for chainId=%s", chainId)
		return
	}

	blocNumberStats := stats.BenchmarkBlockNumber(rpcs, chainId)
	pointsCount, err := influx.WritePoints(influxClient, "stats-block-number-1", blocNumberStats)

	if err != nil {
		fmt.Printf("failed to write points to influx\n")
	}

	fmt.Printf("wrote %d points to influx\n", pointsCount)
}

func setInterval(intervalAction func(), interval time.Duration) {
	ticker := time.Tick(interval)

	for range ticker {
		intervalAction()
	}
}

func MustGetEnv(envName string) string {
	envValue, ok := os.LookupEnv(envName)

	if !ok {
		panic(fmt.Sprintf("failed to get name=%s env", envName))
	}

	return envValue
}
