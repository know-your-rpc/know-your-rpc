package server

import (
	"koonopek/know_your_rpc/server/eth_client"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

type ServerContext struct {
	InfluxClient *influxdb3.Client
	EthClient    *eth_client.EthClient
}
