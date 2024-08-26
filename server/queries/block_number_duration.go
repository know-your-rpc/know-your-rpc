package queries

import (
	"context"
	"fmt"
	"koonopek/know_your_rpc/common/types"
	"koonopek/know_your_rpc/server/server"
	"math"
	"net/http"
	"strings"
	"text/template"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
	"github.com/elliotchance/pie/v2"
)

const DEFAULT_AGGREGATION = "avg"
const MAX_PERIOD = 3600 * 7 * 24

type blockNumberDurationQueryTemplate struct {
	From    int    `validate:"required,number,gt=0"`
	To      int    `validate:"required,number,gt=0"`
	BinTime int    `validate:"required,number,lt=10000,gt=0"`
	ChainId string `validate:"required,number,gt=0"`
	RpcUrls string `validate:"required"`
}

func CreateBlockNumberDurationQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	queryTemplate, err := template.New("query").Parse(`SELECT date_bin(INTERVAL '{{.BinTime}} seconds', time) as _time, max("wholeRequestDuration") as maxduration, avg("wholeRequestDuration") as avgduration , "rpcUrl"
				FROM "blockNumber"
				WHERE
				time >= {{.From}}::TIMESTAMP
				AND
				time <= {{.To}}::TIMESTAMP
				AND
				"isError" = 'false'
				AND
				"rpcUrl" IN ({{.RpcUrls}})
				AND
				"chainId" = '{{.ChainId}}'
				GROUP BY 1, "rpcUrl"
				ORDER BY 1 DESC;`)

	if err != nil {
		panic("failed to create query template")
	}
	// on writer get all buckets and merge duplicates on chain id
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		queryParams := r.URL.Query()

		from, to, binTime, chainId, shouldReturn := ParseBasicQueryParams(queryParams, w)
		if shouldReturn {
			return
		}

		rpcUrls, shouldReturn := GetAuthorizedRpcUrls(r, w, chainId)
		if shouldReturn {
			return
		}

		aggr := GetQueryParam(queryParams, "aggr", DEFAULT_AGGREGATION)

		queryTemplateInput := blockNumberDurationQueryTemplate{
			From:    from,
			To:      to,
			BinTime: binTime,
			ChainId: chainId,
			RpcUrls: strings.Join(pie.Map(rpcUrls, func(info types.RpcInfo) string { return fmt.Sprintf("'%s'", info.URL) }), ","),
		}

		queryBuffer, err := PopulateQueryTemplate(queryTemplateInput, queryTemplate)
		if err != nil {
			fmt.Printf("failed to populate query template error=%s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		queryIterator, err := serverContext.InfluxClient.Query(context.Background(), queryBuffer.String(), influxdb3.WithDatabase("stats-block-number"))

		if err != nil {
			fmt.Printf("failed to get response from influx error=%s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		output := CollectPerRpcResponseToChartData(queryIterator, func(value map[string]interface{}) float64 {
			var y float64

			if aggr == "max" {
				y = value["maxduration"].(float64)
			} else {
				y = value["avgduration"].(float64)
			}

			y = math.Round(CapValue(y, 0, 3000))
			return y
		})

		WriteHttpResponse(output, w)
	}
}
