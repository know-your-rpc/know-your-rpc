package queries

import (
	"context"
	"fmt"
	"koonopek/know_your_rpc/server/server"
	"koonopek/know_your_rpc/writer/utils"
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
	// authorize
	// get bucket and use WHERE rpcUrl in ""
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

		userAddress, shouldReturn := getAuthorizedBucketKey(r, w)
		if shouldReturn {
			return
		}
		rpcUrls, err := server.ReadRpcUrlsForUser(userAddress, chainId)

		if err != nil {
			fmt.Printf("failed to read rpc urls for user userAddress=%s chainId=%s error=%s\n", userAddress, chainId, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		aggr := GetQueryParam(queryParams, "aggr", DEFAULT_AGGREGATION)

		queryTemplateInput := blockNumberDurationQueryTemplate{
			From:    from,
			To:      to,
			BinTime: binTime,
			ChainId: chainId,
			RpcUrls: strings.Join(pie.Map(rpcUrls, func(info utils.RpcInfo) string { return fmt.Sprintf("'%s'", info.URL) }), ","),
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

func getAuthorizedBucketKey(r *http.Request, w http.ResponseWriter) (string, bool) {
	if r.Header.Get("Authorization") != "" {
		splitted := strings.Split(r.Header.Get("Authorization"), "#")
		if len(splitted) != 2 {
			fmt.Printf("wrong formatted authorization header authorization_header=%s", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusBadRequest)
			return "", true
		}
		signature := splitted[0]
		msg := splitted[1]
		signer, err := ExtractSigner(signature, msg)
		if err != nil {
			fmt.Printf("failed to extract signer")
			w.WriteHeader(http.StatusUnauthorized)
			return "", true
		}

		return signer, false
	}
	return "public", false
}
