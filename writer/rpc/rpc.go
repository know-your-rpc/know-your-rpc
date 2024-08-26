package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const RPC_CALL_TIMEOUT = 2 * time.Second

type RpcCallRequest struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      string   `json:"id"`
}

type RpcCallResult struct {
	Result               []byte
	WholeRequestDuration int64
}

func makeRpcCall(method string, params []string) RpcCallRequest {
	return RpcCallRequest{
		Jsonrpc: "2.0",
		Id:      "1",
		Method:  method,
		Params:  params,
	}
}

type rpcCallResponseEnvelope struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      string          `json:"id"`
	Result  json.RawMessage `json:"result"`
}

var rpcHttpClient = http.Client{
	Timeout: RPC_CALL_TIMEOUT,
}

// TODO: use better tracking with http request context
// https://pkg.go.dev/net/http/httptrace
func RpcCall(rpcUrl string, method string, params []string) (RpcCallResult, error) {
	rpcRequest := makeRpcCall(method, params)

	reqBody, err := json.Marshal(rpcRequest)

	if err != nil {
		return RpcCallResult{}, err
	}

	startTime := time.Now()

	resp, err := rpcHttpClient.Post(rpcUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return RpcCallResult{}, err
	}

	respBody, err := io.ReadAll(resp.Body)
	duration := time.Since(startTime).Milliseconds()

	if resp.StatusCode >= 300 {
		return RpcCallResult{}, fmt.Errorf("http error status code=%d", resp.StatusCode)
	}

	if err != nil {
		return RpcCallResult{}, err
	}

	rpcEnvelope := &rpcCallResponseEnvelope{}

	err = json.Unmarshal(respBody, rpcEnvelope)

	if err != nil {
		return RpcCallResult{}, err
	}

	if rpcEnvelope.Id != rpcRequest.Id {
		return RpcCallResult{}, fmt.Errorf("response id(%s) is not matching request id(%s)", rpcEnvelope.Id, rpcRequest.Id)
	}

	return RpcCallResult{
		Result:               rpcEnvelope.Result,
		WholeRequestDuration: duration,
	}, nil
}
