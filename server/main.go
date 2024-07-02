package main

import (
	"fmt"
	"koonopek/know_your_rpc/server/queries"
	"koonopek/know_your_rpc/server/server"
	"koonopek/know_your_rpc/writer/influx"
	"koonopek/know_your_rpc/writer/utils"
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

	handler.Handle("/", http.FileServer(http.Dir("static")))

	addr := fmt.Sprintf(":%s", utils.GetEnvOrDefault("PORT", "8080"))
	fmt.Printf("listening for http connections on %s \n", addr)
	http.ListenAndServe(addr, handler)
}
