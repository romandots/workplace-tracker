package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/example/workplace-tracker/internal/app"
	"github.com/example/workplace-tracker/internal/handlers"
)

func main() {
	ctx := context.Background()
	a, err := app.New(ctx)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer a.Close()

	env := handlers.NewEnv(a)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/qr", env.QRHandler)
	mux.HandleFunc("/checkin", env.CheckInHandler)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("server started on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
