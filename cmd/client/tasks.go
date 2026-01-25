package main

import (
	"sync"
)

type State struct {
	uid  string
	sign int
}

func FindDistantSynonyms(runner QueryRunner, start string, wanted int, maxDepth int) []string {
	return findRelated(runner, start, wanted, maxDepth, 1)
}

func FindDistantAntonyms(runner QueryRunner, start string, wanted int, maxDepth int) []string {
	return findRelated(runner, start, wanted, maxDepth, -1)
}

func findRelated(runner QueryRunner, start string, wanted int, maxDepth int, targetSign int) []string {
	cache := make(map[string][]RelResponse)
	// visited := make(map[string]bool)
	resultSet := make(map[string]struct{})

	// visited[start] = true

	currLayer := []State{{uid: start, sign: 1}}

	const batchSize = 100

	for depth := 0; depth <= maxDepth; depth++ {
		if len(currLayer) == 0 {
			break
		}

		uniqueIds := make(map[string]struct{})
		for _, s := range currLayer {
			if _, ok := cache[s.uid]; !ok {
				uniqueIds[s.uid] = struct{}{}
			}
		}

		toFetch := make([]string, 0, len(uniqueIds))
		for id := range uniqueIds {
			toFetch = append(toFetch, id)
		}

		if len(toFetch) > 0 {
			var wg sync.WaitGroup
			var mu sync.Mutex

			for i := 0; i < len(toFetch); i += batchSize {
				wg.Add(1)
				go func(offset int) {
					defer wg.Done()
					end := min(offset+batchSize, len(toFetch))

					chunk := toFetch[offset:end]

					nodes, _ := runner.getByIds(chunk, len(chunk), 0)

					mu.Lock()
					for _, n := range nodes {
						cache[n.Uid] = n.Rel
					}
					mu.Unlock()
				}(i)
			}
			wg.Wait()
		}

		if depth == wanted {
			for _, s := range currLayer {
				if s.uid != start && s.sign == targetSign {
					resultSet[s.uid] = struct{}{}
				}
			}
		}

		if depth == maxDepth {
			break
		}

		nextLayer := make([]State, 0, len(currLayer)*2)
		nextLayerMap := make(map[State]bool)

		for _, s := range currLayer {
			edges, ok := cache[s.uid]
			if !ok {
				continue
			}

			for _, e := range edges {
				if e.RelationLabel != "synonym" && e.RelationLabel != "antonym" {
					continue
				}

				// if visited[e.Uid] {
				// 	continue
				// }

				nextSign := s.sign
				if e.RelationLabel == "antonym" {
					nextSign *= -1
				}

				nextState := State{uid: e.Uid, sign: nextSign}

				if !nextLayerMap[nextState] {
					// visited[e.Uid] = true
					nextLayerMap[nextState] = true
					nextLayer = append(nextLayer, nextState)
				}
			}
		}

		currLayer = nextLayer
	}

	out := make([]string, 0, len(resultSet))
	for k := range resultSet {
		out = append(out, k)
	}
	return out
}
