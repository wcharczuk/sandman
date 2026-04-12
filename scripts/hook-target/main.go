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
	flagSkipLogs   = flag.Bool("skip-logs", false, "If we should suppress per-request logging output")
	flagListenAddr = flag.String("listen-addr", "127.0.0.1:8080", "The listen address (defers to the `LISTEN_ADDR` environment variable)")
)

func listenAddr() string {
	if value := os.Getenv("LISTEN_ADDR"); value != "" {
		return value
	}
	return *flagListenAddr
}

func main() {
	flag.Parse()
	addr := listenAddr()
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
	slog.Info("listening on", slog.String("listen_addr", addr))
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("http server failed", slog.Any("err", err))
		os.Exit(1)
	}
}
