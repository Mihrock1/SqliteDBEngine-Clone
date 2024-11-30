package storage

import (
	"errors"
	"golang.org/x/exp/constraints"
	"math"
)

// T is constrained to ordered types
type T interface {
	int | string
}

//func defaultValue[T any]() T {
//	var defaultVal T
//	return defaultVal
//}

type Node[T constraints.Ordered] struct {
	t    int        // Minimum degree (defines the range for number of keys)
	keys []T        // A slice of keys
	n    int        // Current number of keys
	C    []*Node[T] // A slice of child pointers
	leaf bool       // Is true when node is leaf. Otherwise, false
}

func newNode[T constraints.Ordered](t int, leaf bool) *Node[T] {
	return &Node[T]{
		t:    t,
		keys: make([]T, 2*t-1),
		n:    0,
		C:    make([]*Node[T], 2*t),
		leaf: leaf,
	}
}

func (node *Node[T]) insertNonFull(key T) {
	i := node.n - 1
	if node.leaf {
		// Fixed the condition: i >= 0 instead of i <= 0
		for i >= 0 && node.keys[i] > key {
			node.keys[i+1] = node.keys[i]
			i--
		}
		node.keys[i+1] = key
		node.n++
	} else {
		for i >= 0 && node.keys[i] > key {
			i--
		}
		if node.C[i+1].n == 2*node.t-1 {
			node.splitChild(i+1, node.C[i+1])

			if node.keys[i+1] < key {
				i++
			}
		}
		node.C[i+1].insertNonFull(key)
	}
}

// Updated splitChild to use the generic type parameter
func (node *Node[T]) splitChild(i int, y *Node[T]) {
	z := newNode[T](y.t, y.leaf)
	z.n = y.t - 1

	for j := 0; j < y.t-1; j++ {
		z.keys[j] = y.keys[j+y.t]
	}

	if !y.leaf {
		for j := 0; j < y.t; j++ {
			z.C[j] = y.C[j+y.t]
		}
	}

	j := node.n
	for j > i {
		node.C[j+1] = node.C[j]
		j--
	}
	node.C[i+1] = z

	j = node.n - 1
	for j >= i { // Fixed condition to include i
		node.keys[j+1] = node.keys[j]
		j--
	}
	node.keys[i] = y.keys[y.t-1] // Fixed index
	node.n++

	y.n = y.t - 1
}

func (node *Node[T]) traverseRec(keys []T) {
	for i := 0; i < node.n; i++ {
		if !node.leaf {
			node.C[i].traverseRec(keys)
		}
		keys = append(keys, node.keys[i])
	}
	if !node.leaf {
		node.C[node.n].traverseRec(keys)
	}
}

func (node *Node[T]) searchRec(key T) (error, *Node[T], int) {
	i := 0
	for i < node.n {
		if key > node.keys[i] {
			i++
			continue
		} else if key == node.keys[i] {
			return nil, node, i
		} else {
			break
		}
	}
	if node.leaf {
		return errors.New("key does not exist in btree"), nil, -1
	} else {
		return node.C[i].searchRec(key)
	}
}

func (node *Node[T]) deleteRec(key T) (error, *Node[T], int, bool) {
	i := 0
	for i < node.n {
		if key > node.keys[i] {
			i++
			continue
		} else if key == node.keys[i] {
			return nil, node, i, false
		} else {
			break
		}
	}
	if node.leaf {
		return errors.New("key does not exist in btree"), nil, -1, false
	} else {
		err, foundNode, j, res := node.C[i].deleteRec(key)
		if err != nil {
			return err, foundNode, j, res
		} else if res == true {
			if foundNode.n < foundNode.t-1 {
				if node.C[i+1] != nil {
					node.fixChildUnderflow(i, foundNode, node.C[i+1], true)
				} else if node.C[i-1] != nil {
					node.fixChildUnderflow(i, foundNode, node.C[i-1], false)
				} else {
					err = errors.New("no sibling to borrow keys from")
				}
			}
			return err, foundNode, j, res
		} else {
			if foundNode.leaf {
				if foundNode.n > foundNode.t-1 {
					err = foundNode.deleteFromLeaf(j)
				} else {
					err = node.deleteFromLeafWithTMinus1Keys(i, foundNode, j)
				}
			} else {
				var parent, child *Node[T]
				var val T
				if foundNode.C[j+1] != nil {
					parent, child, val = foundNode.findSmallestSubtreeKey(foundNode.C[j+1])
				} else if foundNode.C[j] != nil {
					parent, child, val = foundNode.findLargestSubtreeKey(foundNode.C[j])
				}
				foundNode.keys[j] = val
				if child.n > child.t-1 {
					err = child.deleteFromLeaf(0)
				} else {
					err = parent.deleteFromLeafWithTMinus1Keys(parent.n, child, child.n-1)
				}
			}

			if err != nil {
				return err, foundNode, j, false
			} else {
				return nil, foundNode, j, true
			}
		}
	}
}

func (node *Node[T]) findSmallestSubtreeKey(child *Node[T]) (*Node[T], *Node[T], T) {
	if child.C[0] != nil {
		return child.findSmallestSubtreeKey(child.C[0])
	} else {
		return node, child, child.keys[0]
	}
}

func (node *Node[T]) findLargestSubtreeKey(child *Node[T]) (*Node[T], *Node[T], T) {
	if child.C[child.n] != nil {
		return child.findLargestSubtreeKey(child.C[child.n])
	} else {
		return node, child, child.keys[child.n-1]
	}
}

func (node *Node[T]) fixChildUnderflow(i int, foundNode *Node[T], sibling *Node[T], rightSibling bool) {
	if rightSibling {
		foundNode.keys[foundNode.n-1] = node.keys[i]
		if sibling.n > sibling.t-1 {
			node.keys[i] = sibling.keys[0]
			for k := 0; k < sibling.n-1; k++ {
				sibling.keys[k] = sibling.keys[k+1]
			}
			sibling.n--
		} else {
			node.keys[i] = math.MinInt
			node.n--
			node.C[i+1] = nil
			nInit := foundNode.n
			for k := 0; k < sibling.n-1; k++ {
				foundNode.keys[k+nInit] = sibling.keys[k]
				foundNode.n++
			}
			sibling = nil
		}
	} else {
		foundNode.keys[0] = node.keys[i-1]
		if sibling.n > sibling.t-1 {
			node.keys[i-1] = sibling.keys[sibling.n-1]
			sibling.n--
		} else {
			for k := i - 1; k < node.n-1; k++ {
				node.keys[k] = node.keys[k+1]
				node.C[k] = node.C[k+1]
			}
			node.C[node.n-1] = node.C[node.n]
			node.C[node.n] = nil
			node.keys[node.n-1] = math.MinInt
			node.n--

			nInit := foundNode.n
			for k := sibling.n - 1; k >= 0; k-- {
				foundNode.keys[k+nInit] = foundNode.keys[k]
				foundNode.keys[k] = sibling.keys[k]
				foundNode.n++
			}
			sibling = nil
		}
	}
}

func (node *Node[T]) deleteFromLeaf(i int) error {
	//toDelete := node.keys[i]
	for j := i; j < node.n-1; j++ {
		node.keys[j] = node.keys[j+1]
	}
	node.n--
	return nil
}

func (node *Node[T]) deleteFromLeafWithTMinus1Keys(i int, foundNode *Node[T], j int) error {
	//toDelete := foundNode.keys[j]
	if node.C[i+1] != nil {
		for k := j; k < foundNode.n-1; k++ {
			foundNode.keys[k] = foundNode.keys[k+1]
		}
		node.fixChildUnderflow(i, foundNode, node.C[i+1], true)
	} else if node.C[i-1] != nil {
		for k := j; k > 0; k-- {
			foundNode.keys[k] = foundNode.keys[k-1]
		}
		node.fixChildUnderflow(i, foundNode, node.C[i-1], false)
	} else {
		return errors.New("no sibling to borrow keys from")
	}
	return nil
}
