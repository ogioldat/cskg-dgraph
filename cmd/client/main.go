package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	queryArg := flag.String("query", "", "Query number defined in semanticNames")
	quietArg := flag.Bool("quiet", false, "No write to standard output")
	variablesArg := flag.String("vars", "", "GraphQL variables as a JSON object, e.g. '{\"id\":\"0x123\"}'")
	flag.Parse()

	if *queryArg == "" {
		fmt.Println("Usage: client --query=<query-number> [--vars='{\"id\":\"0x123\"}']")
		os.Exit(1)
	}

	queryNumber := *queryArg

	queryKey, ok := semanticNames[queryNumber]
	if !ok {
		log.Fatalf("Unknown query number %s", queryNumber)
	}

	queries, err := LoadQueries()
	if err != nil {
		log.Fatal("Failed to load query files", err)
	}

	query, ok := queries[queryKey]
	if !ok {
		log.Fatalf("Query %s not found in loaded queries", queryKey)
	}

	vars, err := parseVariables(*variablesArg)
	if err != nil {
		log.Fatalf("Failed to parse variables: %v", err)
	}

	client := NewClient()
	defer client.Close()

	txn := client.NewTxn()
	defer txn.Discard(context.Background())

	resp, err := txn.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	if !*quietArg {
		fmt.Println(string(resp.Json))
	}

}

func parseVariables(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	out := make(map[string]string, len(parsed))
	for k, v := range parsed {
		key := k
		if !strings.HasPrefix(key, "$") {
			key = "$" + key
		}
		out[key] = fmt.Sprint(v)
	}

	return out, nil
}
