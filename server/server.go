package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lateralusd/lateralus/logging"
)

const (
	iface = ":"
)

func StartServer(portNum int, w io.Writer) {
	http.HandleFunc("/", handle)
	addr := fmt.Sprintf("%s%d", iface, portNum)
	go func() {
		http.ListenAndServe(addr, nil)
	}()
}

func handle(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("id")
	logging.Successf("Opened mail from id \"%s\"", user)
}
