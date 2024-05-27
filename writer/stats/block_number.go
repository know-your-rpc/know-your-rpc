package stats

import (
	"fmt"
	"koonopek/know_your_rpc/writer/rpc"
	"koonopek/know_your_rpc/writer/utils"
	"math/big"
	"strconv"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
	"github.com/elliotchance/pie/v2"
)

type PerRpcBlockNumberBenchmark struct {
	WholeRequestDuration int64
	BlockNumber          big.Int
	IsError              bool
	RpcUrl               string
}

type PerChainBlockNumberBenchmarks struct {
	PerRpcBlockNumberBenchmarks []PerRpcBlockNumberBenchmark
	Median                      int64
	Max                         int64
	Min                         int64
	Stddev                      float64
	ChainId                     string
	StartTimestamp              time.Time
}

func (blockNumberBenchmark PerChainBlockNumberBenchmarks) ToPoints() []*influxdb3.Point {
	points := make([]*influxdb3.Point, 0, len(blockNumberBenchmark.PerRpcBlockNumberBenchmarks))
	timestamp := time.Now()

	for _, perRpc := range blockNumberBenchmark.PerRpcBlockNumberBenchmarks {
		blockNumberAsFloat64 := float64(perRpc.BlockNumber.Int64())
		diffWithMedian := blockNumberAsFloat64 - float64(blockNumberBenchmark.Median)

		point := influxdb3.NewPointWithMeasurement("blockNumber").
			SetTag("chainId", blockNumberBenchmark.ChainId).
			SetTag("rpcUrl", perRpc.RpcUrl).
			SetTag("isError", strconv.FormatBool(perRpc.IsError)).
			SetDoubleField("blockNumber", blockNumberAsFloat64).
			SetDoubleField("diffWithMedian", diffWithMedian).
			SetUIntegerField("wholeRequestDuration", uint64(perRpc.WholeRequestDuration))

		points = append(points, point)
	}

	pointMedian := influxdb3.NewPointWithMeasurement("blockNumber_median").
		SetTag("chainId", blockNumberBenchmark.ChainId).
		SetDoubleField("blockNumberMedian", float64(blockNumberBenchmark.Median)).
		SetTimestamp(timestamp)

	points = append(points, pointMedian)

	return points
}

func BenchmarkBlockNumber(rpcs []utils.RpcInfo, chainId string) PerChainBlockNumberBenchmarks {
	fmt.Printf("beginning blockNumber benchmarking for %d rpcs\n", len(rpcs))
	blockNumbersCh := make(chan PerRpcBlockNumberBenchmark, len(rpcs))

	startTimestamp := time.Now()

	for i := range rpcs {
		go func(url string) {
			blockNumbersCh <- benchGetBlocNumber(url)
		}(rpcs[i].URL)
	}

	blockNumberBenchmarks := make([]PerRpcBlockNumberBenchmark, 0, len(rpcs))

	for range rpcs {
		blockNumberBenchmark := <-blockNumbersCh
		blockNumberBenchmarks = append(blockNumberBenchmarks, blockNumberBenchmark)
	}

	blockNumbersWithoutErrors :=
		pie.Map(
			pie.FilterNot(blockNumberBenchmarks, func(b PerRpcBlockNumberBenchmark) bool { return b.IsError }),
			func(b PerRpcBlockNumberBenchmark) int64 { return b.BlockNumber.Int64() },
		)

	median := pie.Median(blockNumbersWithoutErrors)
	max := pie.Max(blockNumbersWithoutErrors)
	min := pie.Min(blockNumbersWithoutErrors)
	stdDev := pie.Stddev(blockNumbersWithoutErrors)
	fmt.Printf("finished blockNumber benchmarking median=%d max=%d min=%d stdDev=%.3f\n", median, max, min, stdDev)

	return PerChainBlockNumberBenchmarks{
		PerRpcBlockNumberBenchmarks: blockNumberBenchmarks,
		Median:                      median,
		Max:                         max,
		Min:                         min,
		Stddev:                      stdDev,
		ChainId:                     chainId,
		StartTimestamp:              startTimestamp,
	}
}

func benchGetBlocNumber(rpcUrl string) PerRpcBlockNumberBenchmark {
	result, err := rpc.RpcCall(rpcUrl, "eth_blockNumber", []string{})

	if err != nil {
		fmt.Printf("rpcUrl=%s failed error=%s\n", rpcUrl, err)
		return PerRpcBlockNumberBenchmark{-1, *big.NewInt(0), true, rpcUrl}
	}

	blockNumberInHex := parseJsonString(result.Result)
	blockNumber, ok := hexToBigInt(blockNumberInHex)

	if !ok {
		return PerRpcBlockNumberBenchmark{-1, *big.NewInt(0), true, rpcUrl}
	}

	fmt.Printf("rpcUrl=%s blockNumber=%d duration=%d\n", rpcUrl, blockNumber, result.WholeRequestDuration)

	return PerRpcBlockNumberBenchmark{
		result.WholeRequestDuration,
		*blockNumber,
		false,
		rpcUrl,
	}
}

func hexToBigInt(with0x []byte) (*big.Int, bool) {
	hexWithout0x := with0x[2:]

	if len(hexWithout0x)%2 == 1 {
		hexWithout0x = append([]byte("0"), hexWithout0x...)
	}

	i := new(big.Int)

	return i.SetString(string(hexWithout0x), 16)
}

// removes "
func parseJsonString(jsonString []byte) []byte {
	return jsonString[1 : len(jsonString)-1]
}