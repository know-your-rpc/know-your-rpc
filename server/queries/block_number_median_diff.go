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

type blockNumberMedianQueryTemplate struct {
	From    int    `validate:"required,number,gt=0"`
	To      int    `validate:"required,number,gt=0"`
	BinTime int    `validate:"required,number,lt=100000,gt=0"`
	ChainId string `validate:"required,number,gt=0"`
	RpcUrls string `validate:"required"`
}

func CreateBlockNumberDiffFromMedianQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	queryTemplate, err := template.New("query").Parse(`SELECT date_bin(INTERVAL '{{.BinTime}} seconds', time) as _time, min("diffWithMedian") as minmedian, max("diffWithMedian") as maxmedian , "rpcUrl"
				FROM "blockNumber"
				WHERE
				time >= {{.From}}::TIMESTAMP
				AND
				time <= {{.To}}::TIMESTAMP
				AND
				"isError" = 'false'
				AND
				"chainId" = '{{.ChainId}}'
				AND
				"rpcUrl" IN ({{.RpcUrls}})
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

		from, to, binTime, chainId, shouldReturn := ParseBasicQueryParams(queryParams, w)
		if shouldReturn {
			return
		}

		rpcUrls, shouldReturn := GetUserRpcUrls(r, w, chainId)
		if shouldReturn {
			return
		}

		queryTemplateInput := blockNumberMedianQueryTemplate{
			From:    from,
			To:      to,
			BinTime: binTime,
			ChainId: chainId,
			RpcUrls: strings.Join(pie.Map(rpcUrls, func(info types.RpcInfo) string { return fmt.Sprintf("'%s'", info.URL) }), ","),
		}

		queryBuffer, err := PopulateQueryTemplate(queryTemplateInput, queryTemplate)
		if err != nil {
			fmt.Printf("%s", err.Error())
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
			y := value["minmedian"].(float64)
			if math.Abs(value["maxmedian"].(float64)) > math.Abs(value["minmedian"].(float64)) {
				y = value["maxmedian"].(float64)
			}
			y = CapValue(y, -30.0, 10)
			return y
		})

		WriteHttpResponse(output, w)
	}
}
