package queries

import (
	"fmt"
	"koonopek/know_your_rpc/server/server"
	"koonopek/know_your_rpc/writer/config"
	"net/http"
	"time"
)

func CreateSupportedChainsQuery(serverContext *server.ServerContext) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("New site visitor timestamp=%d\n", time.Now().Unix())
		WriteHttpResponse(config.SUPPORTED_CHAINS, w)
	}
}
