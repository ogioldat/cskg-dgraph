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

	res := FindDistantSynonyms(runner, "0x61a8b", 4, 4)

	log.Println(res)

	if err != nil {
		log.Fatal(err)
	}
}
