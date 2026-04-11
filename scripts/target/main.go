package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var (
	flagSkipLogs = flag.Bool("skip-logs", false, "If we should suppress per-request logging output")
	flagBindAddr = flag.String("bind-addr", "127.0.0.1:8080", "The bind address (defers to the `BIND_ADDR` environment variable)")
)

func bindAddr() string {
	if value := os.Getenv("BIND_ADDR"); value != "" {
		return value
	}
	return *flagBindAddr
}

func main() {
	flag.Parse()
	addr := bindAddr()
	srv := http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !*flagSkipLogs {
				start := time.Now()
				defer func() {
					slog.Info("http.request",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.Duration("elapsed", time.Since(start)),
					)
				}()
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "OK!")
		}),
	}
	slog.Info("listening on", slog.String("bind_addr", addr))
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("http server failed", slog.Any("err", err))
		os.Exit(1)
	}
}
