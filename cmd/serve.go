package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/mehix/gopher-burrows/internal/burrows"
	"github.com/spf13/cobra"
)

var (
	addr  string
	fPath string
)

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Expose and http server",
	Run: func(cmd *cobra.Command, args []string) {

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		errs := make(chan error, 1)

		manager := startManager(ctx)
		burrowsStream := make(chan burrows.Burrow)

		go func() {
			manager.Load(burrowsStream)

			fmt.Println(manager.CurrentStatus())
		}()

		go func() {
			defer close(burrowsStream)
			b, err := os.ReadFile(fPath)
			if err != nil {
				errs <- err
				return
			}

			var data []burrows.Burrow
			if err := json.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
				errs <- err
				return
			}

			for _, b := range data {
				select {
				case <-ctx.Done():
					return
				case burrowsStream <- b:
				}
			}
		}()

		srvr := &http.Server{
			Addr:         addr,
			BaseContext:  func(_ net.Listener) context.Context { return ctx },
			ReadTimeout:  time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      httpHandler(manager),
		}

		go func() {
			fmt.Printf("Listening on %s...\n", srvr.Addr)
			errs <- srvr.ListenAndServe()
		}()

		select {
		case err := <-errs:
			if err != nil && errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		case <-ctx.Done():
			stop()
		}

		_ = srvr.Shutdown(context.Background())
	},
}

func init() {
	cmdServe.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "HTTP address to listen on")
	cmdServe.Flags().StringVar(&fPath, "path", "data/initial.json", "Load initial burrows data")
}

func startManager(ctx context.Context) burrows.Manager {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mngr := burrows.NewManager(ctx, logger)

	return mngr
}

func httpHandler(manager burrows.Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	mux.HandleFunc("GET /", showStatus(manager))
	mux.HandleFunc("POST /rent", rentBurrow(manager))
	return mux
}

func showStatus(manager burrows.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		burrows := manager.CurrentStatus()

		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(burrows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func rentBurrow(manager burrows.Manager) http.HandlerFunc {
	type Response struct {
		Burrow burrows.Burrow
		Error  string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := manager.Rentout()

		w.Header().Set("Content-type", "application/json")
		if err != nil {
			_ = json.NewEncoder(w).Encode(Response{Error: err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(Response{Burrow: b})
	}
}
