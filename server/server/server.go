package server

import (
	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
	"github.com/go-playground/validator/v10"
)

type ServerContext struct {
	InfluxClient *influxdb3.Client
	Validator    *validator.Validate
}
