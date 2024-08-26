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

const MAX_REQUEST_DURATION = 1_000
const MAX_BLOCK_DIFF = 10

type topTenErrorRateQueryTemplate struct {
	From    int    `validate:"required,number,gt=0"`
	To      int    `validate:"required,number,gt=0"`
	BinTime int    `validate:"required,number,gt=0"`
	ChainId string `validate:"required,number,gt=0"`
	RpcUrls string `validate:"required"`
}

type TopTenRpcStats struct {
	RpcUrl             string  `json:"rpcUrl"`
	ErrorRate          float64 `json:"errorRate"`
	AvgDiffFromMedian  float64 `json:"avgDiffFromMedian"`
	AvgRequestDuration float64 `json:"avgRequestDuration"`
}

func CreateTopRpcsQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	queryTemplate, err := template.New("query").Parse(`SELECT date_bin_gapfill(INTERVAL '{{.BinTime}} seconds', time) as _time, sum("isError"::BOOLEAN::DOUBLE) as errors, count(*) as all, avg("diffWithMedian") as avgdiff, avg("wholeRequestDuration") as avgduration, "rpcUrl" FROM "blockNumber"
				WHERE
				time >= {{.From}}::TIMESTAMP
				AND
				time <= {{.To}}::TIMESTAMP
				AND
				"chainId" = '{{.ChainId}}'
				AND
				"rpcUrl" IN ({{.RpcUrls}})
				GROUP BY 1, "rpcUrl"
				ORDER BY 1, errors, avgduration DESC;`)

	if err != nil {
		panic("failed to create query template")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		queryParams := r.URL.Query()

		getAll := GetQueryParam(queryParams, "all", "false") == "true"

		from, to, _, chainId, shouldReturn := ParseBasicQueryParams(queryParams, w)
		if shouldReturn {
			return
		}

		rpcUrls, shouldReturn := GetAuthorizedRpcUrls(r, w, chainId)
		if shouldReturn {
			return
		}

		binTime := to - from

		queryTemplateInput := topTenErrorRateQueryTemplate{
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

		output := make([]TopTenRpcStats, 0, 100)

		deDupMap := make(map[string]bool)

		for queryIterator.Next() {
			value := queryIterator.Value()

			// it might be that rpcUrl is already in the S3 but no data exists for it
			_, ok := value["avgdiff"].(float64)
			if !ok {
				continue
			}

			avgDiffFromMedian := value["avgdiff"].(float64)
			avgRequestDuration := value["avgduration"].(float64)

			rpcUrl := value["rpcUrl"].(string)

			// we skip value of limts
			if !getAll {
				if pie.Abs(avgDiffFromMedian) > MAX_BLOCK_DIFF || avgRequestDuration > MAX_REQUEST_DURATION {
					continue
				}
			}

			// that solution is far from ideal, but good enough it exists because of the some bug in the query/influx
			if _, ok := deDupMap[rpcUrl]; ok {
				continue
			}
			deDupMap[rpcUrl] = true

			output = append(output, TopTenRpcStats{
				RpcUrl:             rpcUrl,
				ErrorRate:          value["errors"].(float64) / float64(value["all"].(int64)) * 100.0,
				AvgDiffFromMedian:  avgDiffFromMedian,
				AvgRequestDuration: avgRequestDuration,
			})
		}

		sortedResult := pie.SortUsing(output, func(a TopTenRpcStats, b TopTenRpcStats) bool {
			if math.Abs(a.ErrorRate-b.ErrorRate) > 0.01 {
				return a.ErrorRate < b.ErrorRate
			}

			if math.Abs(a.AvgDiffFromMedian-b.AvgDiffFromMedian) > 0.5 {
				return a.AvgDiffFromMedian < b.AvgDiffFromMedian
			}

			return a.AvgRequestDuration < b.AvgRequestDuration
		})

		WriteHttpResponse(sortedResult, w)
	}
}
