package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/dgraph-io/dgo/v250"
)

type QueryRunner struct {
	conn     *dgo.Dgraph
	queryMap QueryMap
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

func (q *QueryRunner) getByIds(ids []string, first int, offset int) ([]ConceptResponse, error) {
	txn := q.conn.NewTxn()
	defer txn.Discard(context.Background())

	query := q.queryMap[Q01AllByIDs]

	vars := map[string]string{
		"$id":     "[" + strings.Join(ids, ",") + "]",
		"$first":  strconv.Itoa(first),
		"$offset": strconv.Itoa(offset),
	}
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

	return r.Q, nil
}

func (q *QueryRunner) getByLabel(label string) ([]ConceptResponse, error) {
	txn := q.conn.NewTxn()
	defer txn.Discard(context.Background())

	query := q.queryMap[Q19CustomGetByLabels]

	vars := map[string]string{"$label": label}
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

	return r.Q, err
}
