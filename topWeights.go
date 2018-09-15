package main

import (
	"sort"
)

type WeightScore struct {
	Score      int
	Generation int
	Weights    [][][]float64
}

type TopWeights struct {
	Limit  int
	Scores []WeightScore
}

func (tw *TopWeights) Append(ws WeightScore) {
	newScores := append(tw.Scores, ws)
	sort.Slice(newScores, func(i, j int) bool {
		return newScores[i].Score > newScores[j].Score
	})
	if len(newScores) > tw.Limit {
		newScores = newScores[:tw.Limit]
	}

	tw.Scores = newScores
}
