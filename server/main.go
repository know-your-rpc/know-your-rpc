package main

import (
	"fmt"
	"koonopek/know_your_rpc/common/influx"
	"koonopek/know_your_rpc/common/utils"
	"koonopek/know_your_rpc/server/queries"
	"koonopek/know_your_rpc/server/server"
	"net/http"
)

func main() {
	handler := http.NewServeMux()

	influxClient, err := influx.MakeInfluxDbClient()

	if err != nil {
		panic(err)
	}

	serverContext := &server.ServerContext{
		InfluxClient: influxClient,
	}

	handler.HandleFunc("/api/stats/block-numbers/median-diff", queries.CreateBlockNumberDiffFromMedianQuery(serverContext))

	handler.HandleFunc("/api/stats/block-numbers/duration", queries.CreateBlockNumberDurationQuery(serverContext))

	handler.HandleFunc("/api/stats/block-numbers/error-rate", queries.CreateBlockNumberErrorRateQuery(serverContext))

	handler.HandleFunc("/api/stats/top-rpcs", queries.CreateTopRpcsQuery(serverContext))

	handler.HandleFunc("/api/supported-chains", queries.CreateSupportedChainsQuery(serverContext))

	handler.HandleFunc("/api/custom-rpc/add", queries.CreateCustomRpcAddQuery(serverContext))

	handler.HandleFunc("/api/custom-rpc/remove", queries.CreateCustomRpcRemoveQuery(serverContext))

	handler.Handle("/", http.FileServer(http.Dir("static")))

	addr := fmt.Sprintf(":%s", utils.GetEnvOrDefault("PORT", "8080"))
	fmt.Printf("listening for http connections on %s \n", addr)
	http.ListenAndServe(addr, handler)
}
