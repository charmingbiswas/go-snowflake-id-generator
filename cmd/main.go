package main

import (
	"encoding/json"
	"fmt"
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
		parsed := snowflake.Parse(id)
		fmt.Printf("Generated new snowflake - ID: %d | Time: %s | Node: %d | Sequence: %d\n", id, parsed.TimeStamp.Format(time.RFC3339Nano), parsed.MachineId, parsed.Sequence)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":     true,
			"snowflakeId": id,
		})
	})

	log.Println("Starting server on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
