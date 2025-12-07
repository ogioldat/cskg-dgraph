package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/dgraph-io/dgo/v250"
)

type QueryRunner struct {
	conn *dgo.Dgraph
}

type RelResponse struct {
	ConceptResponse
	EdgeId        string `json:"rel|edge_id,omitempty"`
	RelationLabel string `json:"rel|label,omitempty"`
	Relation      string `json:"rel|relation,omitempty"`
	Sentence      string `json:"rel|sentence,omitempty"`
	Source        string `json:"rel|source,omitempty"`
}

type ConceptResponse struct {
	Uid          string        `json:"uid,omitempty"`
	Uri          string        `json:"uri,omitempty"`
	Label        string        `json:"label,omitempty"`
	Rel          []RelResponse `json:"rel,omitempty"`
	IngoingCnt   int           `json:"ingoing_cnt,omitempty"`
	OutgoingCnt  int           `json:"outgoing_cnt,omitempty"`
	NeighborsCnt int           `json:"neighbors_cnt,omitempty"`
	DType        []string      `json:"dgraph.type,omitempty"`
}

type ResponseRoot struct {
	Q []ConceptResponse `json:"q"`
}

func (q *QueryRunner) getById(id string) (*[]ConceptResponse, error) {
	log.Println("Querying by ID", id)

	txn := q.conn.NewTxn()
	defer txn.Discard(context.Background())

	query := `query all($id: string, $first: int = 10000, $offset: int = 0) {
  q(func: uid($id)) {
    uid
    label
    rel_label
    rel @facets {
      uid
      uri
      label
    }
  }
}`
	vars := map[string]string{"$id": id}
	resp, err := txn.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		log.Fatalln("Txn failed", err)
		return nil, err
	}

	var r ResponseRoot
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatalln("Unmarshal failed", err)
		return nil, err
	}

	return &r.Q, err
}
