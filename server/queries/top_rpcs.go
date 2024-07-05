package queries

import (
	"context"
	"fmt"
	"koonopek/know_your_rpc/server/server"
	"net/http"
	"text/template"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

const MAX_REQUEST_DURATION = 1_000
const MAX_OUT_OF_SYNC = 1

type topTenErrorRateQueryTemplate struct {
	From    int    `validate:"required,number,gt=0"`
	To      int    `validate:"required,number,gt=0"`
	BinTime int    `validate:"required,number,gt=0"`
	ChainId string `validate:"required,number,gt=0"`
}

type TopTenRpcStats struct {
	RpcUrl             string  `json:"rpcUrl"`
	ErrorRate          float64 `json:"errorRate"`
	AvgDiffFromMedian  float64 `json:"avgDiffFromMedian"`
	AvgRequestDuration float64 `json:"avgRequestDuration"`
}

func CreateTopRpcsQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	queryTemplate, err := template.New("query").Parse(`SELECT date_bin(INTERVAL '{{.BinTime}} seconds', time) as _time, sum("isError"::BOOLEAN::DOUBLE) as errors ,  count(*) as all , avg("diffWithMedian") as avgdiff, avg("wholeRequestDuration") as avgduration, "rpcUrl" FROM "blockNumber"
				WHERE
				time >= {{.From}}::TIMESTAMP
				AND
				time <= {{.To}}::TIMESTAMP
				AND
				"chainId" = '{{.ChainId}}'
				GROUP BY 1, "rpcUrl"
				ORDER BY 1, errors, avgduration  DESC;`)

	if err != nil {
		panic("failed to create query template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		queryParams := r.URL.Query()

		from, to, _, chainId, shouldReturn := ParseBasicQueryParams(queryParams, w)
		if shouldReturn {
			return
		}

		binTime := to - from

		queryTemplateInput := topTenErrorRateQueryTemplate{
			From:    from,
			To:      to,
			BinTime: binTime,
			ChainId: chainId,
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

		output := make([]TopTenRpcStats, 0, 100)

		for queryIterator.Next() {
			value := queryIterator.Value()
			avgDiffFromMedian := value["avgdiff"].(float64)
			avgRequestDuration := value["avgduration"].(float64)

			// we skip value of limts
			if avgDiffFromMedian > MAX_OUT_OF_SYNC || avgRequestDuration > MAX_REQUEST_DURATION {
				continue
			}

			output = append(output, TopTenRpcStats{
				RpcUrl:             value["rpcUrl"].(string),
				ErrorRate:          value["errors"].(float64) / float64(value["all"].(int64)) * 100.0,
				AvgDiffFromMedian:  value["avgdiff"].(float64),
				AvgRequestDuration: value["avgduration"].(float64),
			})
		}

		WriteHttpResponse(output, w)
	}
}
