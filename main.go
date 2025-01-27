package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jpappel/penny/api"
	"github.com/jpappel/penny/data"
)

func main() {
	const HOSTNAME = ""
	const PORT = 8080
	addr := fmt.Sprintf("%s:%d", HOSTNAME, PORT)

	mux := api.NewMux()

    data.New("file:data.sqlite3")

	slog.Info(fmt.Sprintf("Starting Penny on %s", addr))
	slog.Info(http.ListenAndServe(addr, mux).Error())
}
