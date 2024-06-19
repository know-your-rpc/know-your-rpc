package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
	"github.com/apache/arrow/go/v15/arrow"

	"github.com/go-playground/validator/v10"
)

type ChartJsPoint struct {
	X int64   `json:"x"`
	Y float64 `json:"y"`
}

type ChartJsDataSet struct {
	Label string         `json:"label"`
	Fill  bool           `json:"fill"`
	Data  []ChartJsPoint `json:"data"`
}

type RpcUrlToChartJsPoints = map[string][]ChartJsPoint

const DEFAULT_INTERVAL = "48"
const MAX_POINTS = 1000
const POINTS_PER_SECOND float64 = 0.36

var Validator = validator.New(validator.WithRequiredStructEnabled())

func CapValue(value float64, min float64, max float64) float64 {
	if value > max {
		return max
	}
	if value < min {
		return min
	}
	return value
}

func GetQueryParam(query url.Values, name string, defaultValue string) string {
	if query.Has(name) {
		defaultValue = query.Get(name)
	}
	return defaultValue
}

func WriteHttpResponse(output interface{}, w http.ResponseWriter) {
	outputBytes, err := json.Marshal(output)

	if err != nil {
		fmt.Printf("failed to marshal response %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(outputBytes)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("failed to write response %s", err.Error())
		return
	}
}

func PopulateQueryTemplate(queryTemplateInput interface{}, queryTemplate *template.Template) (bytes.Buffer, error) {
	var queryBuffer bytes.Buffer
	err := Validator.Struct(queryTemplateInput)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to parse template input error=%s", err.Error())
	}

	err = queryTemplate.Execute(&queryBuffer, queryTemplateInput)

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to execute template error=%s", err.Error())
	}
	return queryBuffer, nil
}

func CollectPerRpcResponseToChartData(queryIterator *influxdb3.QueryIterator, calculateY func(value map[string]interface{}) float64) []ChartJsDataSet {
	blockNumberStats := make(map[string][]ChartJsPoint)

	for queryIterator.Next() {
		value := queryIterator.Value()
		rpcUrl := value["rpcUrl"].(string)
		_, ok := blockNumberStats[rpcUrl]

		if !ok {
			blockNumberStats[rpcUrl] = make([]ChartJsPoint, 0, MAX_POINTS)
		}

		x := value["_time"].(arrow.Timestamp).ToTime(arrow.Nanosecond).UnixMilli()

		y := calculateY(value)

		blockNumberStats[rpcUrl] = append(blockNumberStats[rpcUrl], ChartJsPoint{X: x, Y: y})
	}

	output := make([]ChartJsDataSet, 0, 50)

	for k, v := range blockNumberStats {
		output = append(output, ChartJsDataSet{Label: k, Fill: false, Data: v})
	}
	return output
}
