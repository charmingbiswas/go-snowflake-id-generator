package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmingbiswas/go-snowflake-id-generator/pkg/snowflake"
)

func main() {
	var wg sync.WaitGroup
	server := http.Server{
		Addr:    ":8000",
		Handler: nil,
	}

	machineId, err := snowflake.GetMachineIdFromEnv()
	if err != nil {
		log.Fatalf("invalid machine id %v", err)
	}

	node, err := snowflake.NewNode(machineId)
	if err != nil {
		log.Fatalf("failed to create snowflake node: %v", err)
	}

	http.HandleFunc("GET /snowflake/health", func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /snowflake/generate", func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()
		id, err := node.GenerateSnowflakeId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parsed := snowflake.Parse(id)
		fmt.Printf("Generated new snowflake - ID: %d | Time: %s | Node: %d | Sequence: %d\n", id, parsed.TimeStamp.Format(time.RFC3339Nano), parsed.MachineId, parsed.Sequence)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":     true,
			"snowflakeId": id,
		})
	})

	go func() {
		fmt.Println("Starting server on port 8000")
		fmt.Println(server.ListenAndServe())
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-shutdownCtx.Done()

	fmt.Println("Gracefully shutting down server...")
	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown with error %v", err)
	}

	fmt.Println("Server shutdown complete")

	fmt.Println("Waiting for go routines to shut down")
	wg.Wait()
	fmt.Println("All go-routines stopped. Graceful shutdown complete.")
}
