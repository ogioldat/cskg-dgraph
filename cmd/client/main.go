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
	c := newClient()

	runner := QueryRunner{conn: c}
	resp, err := runner.getById("0x23aa7e")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp)
}
