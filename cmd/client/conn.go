package main

import (
	"log"

	"github.com/dgraph-io/dgo/v250"
	"google.golang.org/grpc"
)

func NewClient() *dgo.Dgraph {
	client, err := dgo.NewClient(
		"localhost:9080",
		dgo.WithGrpcOption(
			grpc.WithInsecure(),
		),
		dgo.WithGrpcOption(
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(100*1024*1024),
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
