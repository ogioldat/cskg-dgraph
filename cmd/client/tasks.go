package main

import (
	"log"
	"math"
	"os"
	"sync"

	"github.com/dominikbraun/graph"

	"github.com/dominikbraun/graph/draw"
)

/*
The dataset contains many synonyms and antonyms (their ids are
"/r/Synonym", "/r/Antonym"). Let's define a distant synonym and
antonym as a node connected by a series of synonym or antonym relations. Synonym of a synonym is a synonym, synonym of an antonym
is an antonym, antonym of an antonym is a synonym, etc. Find all
distant synonyms of a given node at a specified distance (length of the
shortest path), given as a command parameter.

n1 -synonym-> n2 -synonym-> n3 -antonym-> n4 -antonym-> n5 (synonym)
*/

func getByIdsBatched(runner QueryRunner, ids []string, batchSize int) []ConceptResponse {
	numBatches := int(math.Ceil(float64(len(ids)) / float64(batchSize)))

	results := make([][]ConceptResponse, numBatches)

	wg := sync.WaitGroup{}
	wg.Add(numBatches)

	for i := range numBatches {
		go func(i int) {
			log.Println("Fetching batch", i, batchSize)
			defer wg.Done()
			res, _ := runner.getByIds(ids, batchSize, i)

			results[i] = res
		}(i)
	}

	wg.Wait()

	var flatRes []ConceptResponse

	for _, res := range results {
		flatRes = append(flatRes, res...)
	}

	return flatRes
}

func recursiveSearch(
	g graph.Graph[string, string],
	runner QueryRunner,
	uids []string,
	depth int,
	maxDepth int,
) {
	levelNodes := getByIdsBatched(runner, uids, 5)

	var levelUids []string

	for _, node := range levelNodes {
		g.AddVertex(node.Label)

		for _, edge := range node.Rel {
			levelUids = append(levelUids, edge.Uid)

			if edge.RelationLabel != "synonym" && edge.RelationLabel != "antonym" {
				continue
			}

			g.AddVertex(edge.Label)

			g.AddEdge(
				node.Label,
				edge.Label,
				graph.EdgeAttribute("label", edge.RelationLabel),
			)
		}
	}

	if len(levelUids) > 0 {
		recursiveSearch(g, runner, levelUids, depth+1, maxDepth)
	}
}

func FindDistantSynonyms(runner QueryRunner, targetNodeId string) {
	MAX_DEPTH := 1

	g := graph.New(graph.StringHash)
	var depth int

	recursiveSearch(g, runner, []string{targetNodeId}, depth, MAX_DEPTH)

	file, _ := os.Create("./img/distant-synonyms.gv")
	_ = draw.DOT(g, file)

}
