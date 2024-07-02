package queries

import (
	"context"
	"fmt"
	"koonopek/know_your_rpc/server/server"
	"net/http"
	"text/template"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

type blockNumberErrorRateQueryTemplate struct {
	From    int `validate:"required,number,gt=0"`
	To      int `validate:"required,number,gt=0"`
	BinTime int `validate:"required,number,lt=10000,gt=0"`
}

func CreateBlockNumberErrorRateQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	queryTemplate, err := template.New("query").Parse(`SELECT date_bin(INTERVAL '{{.BinTime}} seconds', time) as _time, sum("isError"::BOOLEAN::DOUBLE) as errors ,  count(*) as all , "rpcUrl"
				FROM "blockNumber"
				WHERE
				time >= {{.From}}::TIMESTAMP
				AND
				time <= {{.To}}::TIMESTAMP
				AND
				"chainId" = '1'
				GROUP BY 1, "rpcUrl"
				ORDER BY 1 DESC;`)

	if err != nil {
		panic("failed to create query template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		queryParams := r.URL.Query()

		from, to, binTime, shouldReturn := ParseTimeQuery(queryParams, w)
		if shouldReturn {
			return
		}
		queryTemplateInput := blockNumberErrorRateQueryTemplate{
			From:    from,
			To:      to,
			BinTime: binTime,
		}

		queryBuffer, err := PopulateQueryTemplate(queryTemplateInput, queryTemplate)
		if err != nil {
			fmt.Printf("%s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
		}

		queryIterator, err := serverContext.InfluxClient.Query(context.Background(), queryBuffer.String(), influxdb3.WithDatabase("stats-block-number"))

		if err != nil {
			fmt.Printf("failed to get response from influx error=%s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		output := CollectPerRpcResponseToChartData(queryIterator, func(value map[string]interface{}) float64 {
			return value["errors"].(float64) / float64(value["all"].(int64)) * 100.0
		})

		WriteHttpResponse(output, w)
	}
}