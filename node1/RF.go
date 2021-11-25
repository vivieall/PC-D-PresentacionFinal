package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Forest struct {
	Trees []*Tree
}

func TrainForest(data [][]interface{}, labels []string, samples, features, trees int) *Forest {
	rand.Seed(time.Now().UnixNano())
	forest := &Forest{}
	forest.Trees = make([]*Tree, trees)
	done_flag := make(chan bool)
	mutex := &sync.Mutex{}
	for i := 0; i < trees; i++ {
		go func(x int) {
			fmt.Printf("Entrenando árbol %v \n", x+1)
			forest.Trees[x] = TrainTree(data, labels, samples, features)
			fmt.Printf("Arbol %v está listo\n", x+1)
			mutex.Lock()
			mutex.Unlock()
			done_flag <- true
		}(i)
	}

	for i := 1; i <= trees; i++ {
		<-done_flag
	}

	return forest
}

func (f *Forest) Predicate(input []interface{}) string {
	counter := make(map[string]float64)
	for i := 0; i < len(f.Trees); i++ {
		tree_counter := PredicateTree(f.Trees[i], input)
		total := 0.0
		for _, v := range tree_counter {
			total += float64(v)
		}
		for k, v := range tree_counter {
			counter[k] += float64(v) / total
		}
	}

	max_c := 0.0
	max_label := ""
	for k, v := range counter {
		if v >= max_c {
			max_c = v
			max_label = k
		}
	}
	return max_label
}
