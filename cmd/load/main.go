package main

import (
	"context"
	"log"
	"os"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
)

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}

func loadSchema(c *dgo.Dgraph) {
	// Install a schema into dgraph. Accounts have a `name` and a `balance`.
	schemaBytes, err := os.ReadFile("schema.dql")
	if err != nil {
		log.Fatal(err)
	}

	log.Default().Println(string(schemaBytes))
	err = c.Alter(context.Background(), &api.Operation{
		Schema: string(schemaBytes),
	})
}

func main() {
	c := newClient()
	loadSchema(c)
}
