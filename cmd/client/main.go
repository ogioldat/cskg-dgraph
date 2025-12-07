package main

import (
	"log"
)

func main() {
	queries, err := LoadQueries()
	if err != nil {
		log.Fatal("Failed to load query files", err)
	}

	c := NewClient()

	runner := QueryRunner{conn: c, queryMap: queries}
	nodes, err := runner.getByLabel("synonym")

	FindDistantSynonyms(runner, "0x2f4e0b")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(nodes)
}
