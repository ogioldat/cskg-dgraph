package main

import (
	"log"

	"github.com/dgraph-io/dgo/v250"
)

func newClient() *dgo.Dgraph {
	client, err := dgo.Open("dgraph://localhost:9080")
	if err != nil {
		client.Close()
		log.Fatal(err)
	}

	return client
}

func main() {
	queries, err := LoadQueries()
	if err != nil {
		log.Fatal("Failed to load query files", err)
	}

	c := newClient()

	runner := QueryRunner{conn: c, queryMap: queries}
	resp, err := runner.getByLabel("synonym")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
