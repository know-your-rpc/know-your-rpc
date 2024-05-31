package influx

import (
	"context"
	"fmt"
	"koonopek/know_your_rpc/writer/utils"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

type ToInfluxPoints interface {
	ToPoints() []*influxdb3.Point
}

func MakeInfluxDbClient() (*influxdb3.Client, error) {
	// Create client
	url := utils.MustGetEnv("INFLUXDB_URL")
	token := utils.MustGetEnv("INFLUXDB_TOKEN")

	// Create a new client using an InfluxDB server base URL and an authentication token
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:  url,
		Token: token,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create influx db client err= %s", err)
	}

	return client, nil
}

func WritePoints(client *influxdb3.Client, bucket string, toPoint ToInfluxPoints) (int, error) {
	points := toPoint.ToPoints()
	err := client.WritePoints(context.Background(), points, influxdb3.WithDatabase(bucket))

	return len(points), err
}
