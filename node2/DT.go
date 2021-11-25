package main

import (
	"math"
	"math/rand"
)

const CAT = "cat"
const NUMERIC = "numeric"

type TreeNode struct {
	ColumnNo int
	Value    interface{}
	Left     *TreeNode
	Right    *TreeNode
	Labels   map[string]int
}

type Tree struct {
	Root *TreeNode
}

func getRandomRange(N int, M int) []int {
	tmp := make([]int, N)
	for i := 0; i < N; i++ {
		tmp[i] = i
	}
	for i := 0; i < M; i++ {
		j := i + int(rand.Float64()*float64(N-i))
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}

	return tmp[:M]
}

func getSamples(a [][]interface{}, idx []int) [][]interface{} {
	result := make([][]interface{}, len(idx))
	for i := 0; i < len(idx); i++ {
		result[i] = a[idx[i]]
	}
	return result
}

func getLabels(a []string, idx []int) []string {
	result := make([]string, len(idx))
	for i := 0; i < len(idx); i++ {
		result[i] = a[idx[i]]
	}
	return result
}

func getEntropy(ep_map map[string]float64, total int) float64 {

	for k := range ep_map {
		ep_map[k] = ep_map[k] / float64(total)
	}

	entropy := 0.0
	for _, v := range ep_map {
		entropy += v * math.Log(1.0/v)
	}

	return entropy
}

func getBestGain(data [][]interface{}, c int, samples_labels []string, column_type string, current_entropy float64) (float64, interface{}, int, int) {
	var best_value interface{}
	best_gain := 0.0
	best_total_r := 0
	best_total_l := 0
	uniq_values := make(map[interface{}]int)
	for i := 0; i < len(data); i++ {
		uniq_values[data[i][c]] = 1
	}
	for value := range uniq_values {
		map_l := make(map[string]float64)
		map_r := make(map[string]float64)
		total_l := 0
		total_r := 0
		if column_type == CAT {
			for j := 0; j < len(data); j++ {
				if data[j][c] == value {
					total_l += 1
					map_l[samples_labels[j]] += 1.0
				} else {
					total_r += 1
					map_r[samples_labels[j]] += 1.0
				}
			}
		}
		if column_type == NUMERIC {
			for j := 0; j < len(data); j++ {
				if data[j][c].(float64) <= value.(float64) {
					total_l += 1
					map_l[samples_labels[j]] += 1.0
				} else {
					total_r += 1
					map_r[samples_labels[j]] += 1.0
				}
			}
		}
		p1 := float64(total_r) / float64(len(data))
		p2 := float64(total_l) / float64(len(data))
		new_entropy := p1*getEntropy(map_r, total_r) + p2*getEntropy(map_l, total_l)
		entropy_gain := current_entropy - new_entropy
		if entropy_gain >= best_gain {
			best_gain = entropy_gain
			best_value = value
			best_total_l = total_l
			best_total_r = total_r
		}
	}

	return best_gain, best_value, best_total_l, best_total_r
}

func splitData(data [][]interface{}, column_type string, c int, value interface{}, part_l *[]int, part_r *[]int) {
	if column_type == CAT {
		for j := 0; j < len(data); j++ {
			if data[j][c] == value {
				*part_l = append(*part_l, j)
			} else {
				*part_r = append(*part_r, j)
			}
		}
	}
	if column_type == NUMERIC {
		for j := 0; j < len(data); j++ {
			if data[j][c].(float64) <= value.(float64) {
				*part_l = append(*part_l, j)
			} else {
				*part_r = append(*part_r, j)
			}
		}
	}
}

func generateTree(data [][]interface{}, samples_labels []string, features int) *TreeNode {
	column_count := len(data[0])
	split_count := features
	columns_choosen := getRandomRange(column_count, split_count)

	best_gain := 0.0
	var best_part_l []int = make([]int, 0, len(data))
	var best_part_r []int = make([]int, 0, len(data))
	var best_total_l int = 0
	var best_total_r int = 0
	var best_value interface{}
	var best_column int
	var best_column_type string

	current_entropy_map := make(map[string]float64)
	for i := 0; i < len(samples_labels); i++ {
		current_entropy_map[samples_labels[i]] += 1
	}

	current_entropy := getEntropy(current_entropy_map, len(samples_labels))

	for _, c := range columns_choosen {
		column_type := CAT
		if _, ok := data[0][c].(float64); ok {
			column_type = NUMERIC
		}

		gain, value, total_l, total_r := getBestGain(data, c, samples_labels, column_type, current_entropy)
		if gain >= best_gain {
			best_gain = gain
			best_value = value
			best_column = c
			best_column_type = column_type
			best_total_l = total_l
			best_total_r = total_r
		}
	}
	if best_gain > 0 && best_total_l > 0 && best_total_r > 0 {
		node := &TreeNode{}
		node.Value = best_value
		node.ColumnNo = best_column
		splitData(data, best_column_type, best_column, best_value, &best_part_l, &best_part_r)
		node.Left = generateTree(getSamples(data, best_part_l), getLabels(samples_labels, best_part_l), features)
		node.Right = generateTree(getSamples(data, best_part_r), getLabels(samples_labels, best_part_r), features)
		return node
	}
	return genLeafNode(samples_labels)
}

func genLeafNode(labels []string) *TreeNode {
	counter := make(map[string]int)
	for _, v := range labels {
		counter[v] += 1
	}

	node := &TreeNode{}
	node.Labels = counter
	return node
}

func TrainTree(inputs [][]interface{}, labels []string, samples, features int) *Tree {

	data := make([][]interface{}, samples)
	samples_labels := make([]string, samples)
	for i := 0; i < samples; i++ {
		j := int(rand.Float64() * float64(len(inputs)))
		data[i] = inputs[j]
		samples_labels[i] = labels[j]
	}

	tree := &Tree{}
	tree.Root = generateTree(data, samples_labels, features)

	return tree
}

func predicate(node *TreeNode, input []interface{}) map[string]int {
	if node.Labels != nil { //leaf node
		return node.Labels
	}

	c := node.ColumnNo
	value := input[c]

	switch value.(type) {
	case float64:
		if value.(float64) <= node.Value.(float64) && node.Left != nil {
			return predicate(node.Left, input)
		} else if node.Right != nil {
			return predicate(node.Right, input)
		}
	case string:
		if value == node.Value && node.Left != nil {
			return predicate(node.Left, input)
		} else if node.Right != nil {
			return predicate(node.Right, input)
		}
	}

	return nil
}

func PredicateTree(tree *Tree, input []interface{}) map[string]int {
	return predicate(tree.Root, input)
}
