package main

import (
	"log"
	"net/http"
	"os"

	"energy-dashboard-go/internal/server"
)

func main() {
	s := server.New()
	addr := ":3000"
	if v := os.Getenv("PORT"); v != "" { addr = ":" + v }
	log.Println("Energy Dashboard (Go) listening on", addr)
	log.Fatal(http.ListenAndServe(addr, s))
}
