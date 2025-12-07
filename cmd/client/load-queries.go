package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

const queryDir = "gql"

type QueryKey string
type QueryMap map[QueryKey]string

const (
	Q01AllByID                 QueryKey = "Q01AllByID"
	Q02CountSuccessors         QueryKey = "Q02CountSuccessors"
	Q03FindPredecessors        QueryKey = "Q03FindPredecessors"
	Q04CountPredecessors       QueryKey = "Q04CountPredecessors"
	Q05FindAllNeighbors        QueryKey = "Q05FindAllNeighbors"
	Q06CountAllNeighbors       QueryKey = "Q06CountAllNeighbors"
	Q07FindGrandchildren       QueryKey = "Q07FindGrandchildren"
	Q08FindGrandparents        QueryKey = "Q08FindGrandparents"
	Q09CountTotalNodes         QueryKey = "Q09CountTotalNodes"
	Q10NodesWithNoSuccessors   QueryKey = "Q10NodesWithNoSuccessors"
	Q11NodesWithNoPredecessors QueryKey = "Q11NodesWithNoPredecessors"
	Q12NodesWithMostNeighbors  QueryKey = "Q12NodesWithMostNeighbors"
	Q13NodesWithSingleNeighbor QueryKey = "Q13NodesWithSingleNeighbor"
	Q15FindSimilarNodes        QueryKey = "Q15FindSimilarNodes"
	Q16CheckShortestPath       QueryKey = "Q16CheckShortestPath"
	Q17FindDistantSynonyms     QueryKey = "Q17FindDistantSynonyms"
	Q18FindDistantAntonyms     QueryKey = "Q18FindDistantAntonyms"
)

var semanticNames = map[string]QueryKey{
	"1":  Q01AllByID,
	"2":  Q02CountSuccessors,
	"3":  Q03FindPredecessors,
	"4":  Q04CountPredecessors,
	"5":  Q05FindAllNeighbors,
	"6":  Q06CountAllNeighbors,
	"7":  Q07FindGrandchildren,
	"8":  Q08FindGrandparents,
	"9":  Q09CountTotalNodes,
	"10": Q10NodesWithNoSuccessors,
	"11": Q11NodesWithNoPredecessors,
	"12": Q12NodesWithMostNeighbors,
	"13": Q13NodesWithSingleNeighbor,
	"15": Q15FindSimilarNodes,
	"16": Q16CheckShortestPath,
	"17": Q17FindDistantSynonyms,
	"18": Q18FindDistantAntonyms,
}

func LoadQueries() (QueryMap, error) {
	queries := make(QueryMap)

	re := regexp.MustCompile(`^(\d+)-.*\.gql$`)

	files, err := ioutil.ReadDir(queryDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read query directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()

		matches := re.FindStringSubmatch(fileName)
		if len(matches) < 2 {
			continue
		}
		fileNumber := matches[1]

		semanticName, exists := semanticNames[fileNumber]
		if !exists {
			fmt.Printf("Warning: Semantic name not defined for file number %s. Skipping.\n", fileNumber)
			continue
		}

		filePath := filepath.Join(queryDir, fileName)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", fileName, err)
		}

		queries[semanticName] = string(content)
	}

	return queries, nil
}
