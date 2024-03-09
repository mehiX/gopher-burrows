package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/mehix/gopher-burrows/cmd"
	"github.com/mehix/gopher-burrows/internal/burrows"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func Run() {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	burrowsStream := make(chan burrows.Burrow)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	manager := burrows.NewManager(ctx, logger)
	var wg sync.WaitGroup
	wg.Add(3)
	for i := range 3 {
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Int63n(3)) * time.Second)
			burrowsStream <- burrows.Burrow{Name: fmt.Sprintf("burrows %d", i+1), Depth: 2.5, Width: 1.2, AgeInMin: 10}
		}()
	}
	go func() {
		wg.Wait()
		close(burrowsStream)
	}()
	manager.Load(burrowsStream)

	log.Println("start gophers")
	for range 5 {
		go func() {
			for {
				time.Sleep(time.Duration(rand.Int63n(5)) * time.Second)
				b, err := manager.Rentout()
				if err != nil {
					log.Println("rentingout", err)
				} else {
					log.Println("success rentout", b)
					return
				}
			}
		}()
	}

	repoTkr := time.NewTicker(10 * time.Second)
	defer repoTkr.Stop()

	for {
		select {
		case <-time.Tick(5 * time.Second):
			burrows := manager.CurrentStatus()
			for _, b := range burrows {
				fmt.Println(b)
			}
		case <-repoTkr.C:
			rep := manager.Report()
			b, _ := json.MarshalIndent(rep, "  ", "  ")
			io.Copy(os.Stdout, bytes.NewReader(b))
			fmt.Println()
		case <-ctx.Done():
			<-manager.Done
			log.Println("Done")
			panic("show traces")
		}
	}
}
