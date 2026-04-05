package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/charmingbiswas/go-snowflake-id-generator/pkg/snowflake"
)

func main() {
	machineId, err := snowflake.GetMachineIdFromEnv()
	if err != nil {
		log.Fatalf("invalid machine id %v", err)
	}

	node, err := snowflake.NewNode(machineId)
	if err != nil {
		log.Fatalf("failed to create snowflake node: %v", err)
	}

	http.HandleFunc("GET /snowflake/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /snowflake/generate", func(w http.ResponseWriter, r *http.Request) {
		id, err := node.GenerateSnowflakeId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":   true,
			"id":        id,
			"timestamp": time.Now().UnixMilli(),
		})
	})
}
