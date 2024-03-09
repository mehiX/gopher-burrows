package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/mehix/gopher-burrows/internal/burrows"
	"github.com/spf13/cobra"
)

var (
	logger        *slog.Logger
	addr          string
	fPath         string
	verbose       bool
	reportingDir  string
	reportingFreq time.Duration
)

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Expose and http server",
	Run: func(cmd *cobra.Command, args []string) {

		logLevel := slog.LevelInfo
		if verbose {
			logLevel = slog.LevelDebug
		}
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		errs := make(chan error, 1)

		// Create manager and load data
		manager := burrows.NewManager(ctx, logger)

		burrowsStream := make(chan burrows.Burrow)

		go func() {
			manager.Load(burrowsStream)

			logger.Debug("manager data loaded", "data", manager.CurrentStatus())
		}()

		go loadInitialData(ctx, burrowsStream, errs)

		go generatePeriodicReports(ctx, manager, errs)

		// Create the HTTP server
		srvr := &http.Server{
			Addr:         addr,
			BaseContext:  func(_ net.Listener) context.Context { return ctx },
			ReadTimeout:  time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      httpHandler(manager),
		}

		go func() {
			logger.Info("http server started", "addr", srvr.Addr)
			errs <- srvr.ListenAndServe()
		}()

		select {
		case err := <-errs:
			if err != nil && errors.Is(err, http.ErrServerClosed) {
				logger.Error(err.Error())
			}
		case <-ctx.Done():
			stop()
		}

		_ = srvr.Shutdown(context.Background())

		<-manager.Done
	},
}

func init() {
	cmdServe.Flags().StringVar(&addr, "addr", "127.0.0.1:8080", "HTTP address to listen on")
	cmdServe.Flags().StringVar(&fPath, "path", "data/initial.json", "Load initial burrows data")
	cmdServe.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable more verbose logging")

	cmdServe.Flags().StringVar(&reportingDir, "repos-dir", "/tmp", "path to write out reports")
	cmdServe.Flags().DurationVar(&reportingFreq, "repos-freq", 10*time.Minute, "frequency for writing out reports")

	cmdServe.Flags().DurationVarP(&burrows.Tact, "tact", "t", time.Minute, "change the speed with which the data is generated")
}

func loadInitialData(ctx context.Context, burrowsStream chan<- burrows.Burrow, errs chan<- error) {
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
}

func generatePeriodicReports(ctx context.Context, manager burrows.Manager, errs chan<- error) {

	tkr := time.NewTicker(reportingFreq)
	defer tkr.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tkr.C:
			fpath := filepath.Join(reportingDir, fmt.Sprintf("%s_%s.txt", "burrows", time.Now().Format("20060102_150405")))
			report := manager.Report()

			f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664)
			if err != nil {
				errs <- err
				return
			}
			defer f.Close()

			w := tabwriter.NewWriter(f, 15, 0, 0, ' ', tabwriter.AlignRight)
			defer w.Flush()

			if err := report.ToTxt(f); err == nil {
				logger.Info("report generated", "filename", fpath)
			}

		}
	}
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
		allowedTime, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		b, err := manager.Rentout(allowedTime)

		w.Header().Set("Content-type", "application/json")
		if err != nil {
			_ = json.NewEncoder(w).Encode(Response{Error: err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(Response{Burrow: b})
	}
}
