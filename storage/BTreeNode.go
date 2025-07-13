package storage

import (
	"errors"
	"golang.org/x/exp/constraints"
	"reflect"
)

//func defaultValue[T any]() T {
//	var defaultVal T
//	return defaultVal
//}

// TODO: check if m is needed here when it already exists in btree
type Node[T constraints.Ordered] struct {
	m      int        // order of BTree Node
	n      int        // Current number of keys
	K      []T        // A slice of keys
	C      []*Node[T] // A slice of child pointers
	isLeaf bool       // Is true when node is isLeaf. Otherwise, false
}

func newNode[T constraints.Ordered](order int, leaf bool) *Node[T] {
	return &Node[T]{
		m:      order,
		K:      make([]T, order-1),
		n:      0,
		C:      make([]*Node[T], order),
		isLeaf: leaf,
	}
}

func (node *Node[T]) insertNonFull(key T) {
	i := node.n - 1
	if node.isLeaf {
		for i >= 0 && node.K[i] > key {
			node.K[i+1] = node.K[i]
			i--
		}
		node.K[i+1] = key
		node.n++
	} else {
		for i >= 0 && node.K[i] > key {
			i--
		}
		if node.C[i+1].n == node.m-1 {
			node.splitChild(i+1, node.C[i+1])

			if node.K[i+1] < key {
				i++
			}
		}
		node.C[i+1].insertNonFull(key)
	}
}

// TODO: check if two params are needed
func (node *Node[T]) splitChild(i int, child *Node[T]) {
	newChild := newNode[T](child.m, child.isLeaf)
	// TODO: update value of n as key in each node is added or deleted to make debugging easier
	//newChild.n = child.m - 1

	for j := 0; j < child.m-1; j++ {
		// move keys from second half of first child to new child
		newChild.K[j] = child.K[j+child.m]
		newChild.n++
		child.n--

		// move child pointers from second half of first child to new child
		newChild.C[j] = child.C[j+child.m]
		child.C[j+child.m] = nil
	}

	// for last child pointer not encountered in above loop
	newChild.C[child.m-1] = child.C[2*child.m-1]
	child.C[2*child.m-1] = nil

	//if !child.isLeaf {
	//	for j := 0; j < child.m; j++ {
	//		newChild.C[j] = child.C[j+child.m]
	//		child.C[j+child.m] = nil
	//	}
	//}

	// moving forward child pointers after current child index by 1 index
	j := node.n
	for j > i {
		node.C[j+1] = node.C[j]
		j--
	}
	// adding address of new child on index next to current child
	node.C[i+1] = newChild

	// moving forward keys after current child/key index by 1 index
	j = node.n - 1
	for j >= i {
		node.K[j+1] = node.K[j]
		j--
	}
	// adding new key on index next to current child/key
	node.K[i] = child.K[child.m-1]
	node.n++
	child.n--

	//nInit := child.n
	//child.n = child.m - 1

	// check if casting is needed for explicitly defined ordered types
	//for k := child.n; k < nInit; k++ {
	//	keyDataType := reflect.TypeOf(child.K[k])
	//	if keyDataType == reflect.TypeOf(0) {
	//		child.K[k] = any(0).(T) // Cast 0 to type T
	//	} else if keyDataType == reflect.TypeOf("") {
	//		child.K[k] = any("").(T) // Cast "" to type T
	//	}
	//}
}

func (node *Node[T]) traverseRec(keys []T) {
	for i := 0; i < node.n; i++ {
		if !node.isLeaf {
			node.C[i].traverseRec(keys)
		}
		keys = append(keys, node.K[i])
	}
	if !node.isLeaf {
		node.C[node.n].traverseRec(keys)
	}
}

func (node *Node[T]) searchRec(key T) (error, *Node[T], int) {
	i := 0
	for i < node.n {
		if key > node.K[i] {
			i++
			continue
		} else if key == node.K[i] {
			return nil, node, i
		} else {
			break
		}
	}
	if node.isLeaf {
		return errors.New("key does not exist in btree"), nil, -1
	} else {
		return node.C[i].searchRec(key)
	}
}

func (node *Node[T]) deleteRec(key T) (error, *Node[T], int, bool) {
	i := 0
	for i < node.n {
		if key > node.K[i] {
			i++
			continue
		} else if key == node.K[i] {
			return nil, node, i, false
		} else {
			break
		}
	}
	if node.isLeaf {
		return errors.New("key does not exist in btree"), nil, -1, false
	} else {
		err, child, j, res := node.C[i].deleteRec(key)
		if err != nil {
			return err, child, j, res
		} else if res == true {
			if child.n < child.m-1 {
				node.fixUnderflow(i, child)
			}
			return nil, node, i, true
		} else {
			if child.isLeaf {
				if child.n > child.m-1 {
					child.deleteFromLeaf(j)
				} else {
					node.deleteFromLeafWithTMinus1Keys(i, child, j)
				}
			} else {
				var searchIn, P, C *Node[T]
				var val T
				var largest bool
				if child.C[j] != nil {
					searchIn = child.C[j]
					P, C, val = child.findLargestKeyInSubtreeRec(searchIn)
					largest = true
				} else if child.C[j+1] != nil {
					searchIn = child.C[j+1]
					P, C, val = child.findSmallestKeyInSubtreeRec(searchIn)
					largest = false
				}
				child.K[j] = val
				if C.n > C.m-1 {
					if largest == true {
						C.deleteFromLeaf(C.n - 1)
					} else {
						C.deleteFromLeaf(0)
					}
				} else {
					if largest == true {
						P.deleteFromLeafWithTMinus1Keys(P.n-1, C, C.n-1)
					} else {
						P.deleteFromLeafWithTMinus1Keys(0, C, 0)
					}

				}
			}
			return nil, node, i, true
		}
	}
}

func (node *Node[T]) findSmallestKeyInSubtreeRec(child *Node[T]) (*Node[T], *Node[T], T) {
	if child.C[0] != nil {
		return child.findSmallestKeyInSubtreeRec(child.C[0])
	} else {
		return node, child, child.K[0]
	}
}

func (node *Node[T]) findLargestKeyInSubtreeRec(child *Node[T]) (*Node[T], *Node[T], T) {
	if child.C[child.n-1] != nil {
		return child.findLargestKeyInSubtreeRec(child.C[child.n-1])
	} else {
		return node, child, child.K[child.n-1]
	}
}

func (node *Node[T]) deleteFromLeaf(i int) {
	for j := i; j < node.n-1; j++ {
		node.K[j] = node.K[j+1]
	}
	keyDataType := reflect.TypeOf(node.K[node.n-1])
	if keyDataType == reflect.TypeOf(0) {
		node.K[node.n-1] = any(0).(T) // Cast 0 to type T
	} else if keyDataType == reflect.TypeOf("") {
		node.K[node.n-1] = any("").(T) // Cast "" to type T
	}
	node.n--
}

func (node *Node[T]) deleteFromLeafWithTMinus1Keys(i int, foundNode *Node[T], j int) {
	for k := j; k < foundNode.n-1; k++ {
		foundNode.K[k] = foundNode.K[k+1]
	}
	keyDataType := reflect.TypeOf(foundNode.K[foundNode.n-1])
	if keyDataType == reflect.TypeOf(0) {
		foundNode.K[foundNode.n-1] = any(0).(T) // Cast 0 to type T
	} else if keyDataType == reflect.TypeOf("") {
		foundNode.K[foundNode.n-1] = any("").(T) // Cast "" to type T
	}
	foundNode.n--

	node.fixUnderflow(i, foundNode)
}

func (node *Node[T]) fixUnderflow(i int, child *Node[T]) {
	if i > 0 && node.C[i-1] != nil {
		// left sibling exists case
		for j := child.n - 1; j >= 0; j-- {
			child.K[j+1] = child.K[j]
		}
		child.K[0] = node.K[i-1]
		child.n++
		sibling := node.C[i-1]

		if sibling.n > sibling.m-1 {
			node.K[i-1] = sibling.K[sibling.n-1]

			P, C, val := child.findLargestKeyInSubtreeRec(sibling.C[sibling.n-1])
			sibling.K[sibling.n-1] = val
			if C.n > C.m-1 {
				C.deleteFromLeaf(C.n - 1)
			} else {
				P.deleteFromLeafWithTMinus1Keys(P.n-1, C, C.n-1)
			}
		} else {
			for k := i - 1; k < node.n-1; k++ {
				node.K[k] = node.K[k+1]
				node.C[k] = node.C[k+1]
			}
			node.C[node.n-1] = node.C[node.n]
			node.C[node.n] = nil

			keyDataType := reflect.TypeOf(node.K[node.n-1])
			if keyDataType == reflect.TypeOf(0) {
				node.K[node.n-1] = any(0).(T) // Cast 0 to type T
			} else if keyDataType == reflect.TypeOf("") {
				node.K[node.n-1] = any("").(T) // Cast "" to type T
			}
			node.n--

			nInit := child.n
			for k := sibling.n - 1; k >= 0; k-- {
				child.K[k+nInit] = child.K[k]
				child.C[k+nInit] = child.C[k]
				child.K[k] = sibling.K[k]
				child.C[k] = sibling.C[k]
				child.n++
			}
		}
	} else if node.C[i+1] != nil {
		// right sibling exists case
		child.K[child.n] = node.K[i]
		child.n++
		sibling := node.C[i+1]

		if sibling.n > sibling.m-1 {
			node.K[i] = sibling.K[0]
			for k := 0; k < sibling.n-1; k++ {
				sibling.K[k] = sibling.K[k+1]
			}

			keyDataType := reflect.TypeOf(sibling.K[sibling.n-1])
			if keyDataType == reflect.TypeOf(0) {
				sibling.K[sibling.n-1] = any(0).(T) // Cast 0 to type T
			} else if keyDataType == reflect.TypeOf("") {
				sibling.K[sibling.n-1] = any("").(T) // Cast "" to type T
			}
			sibling.n--
		} else {
			k := i
			for k < node.n-1 {
				node.K[k] = node.K[k+1]
				node.C[k+1] = node.C[k+2]
				k++
			}
			node.C[node.n] = nil

			keyDataType := reflect.TypeOf(node.K[k])
			if keyDataType == reflect.TypeOf(0) {
				node.K[k] = any(0).(T) // Cast 0 to type T
			} else if keyDataType == reflect.TypeOf("") {
				node.K[k] = any("").(T) // Cast "" to type T
			}
			node.n--

			nInit := child.n
			for k := 0; k < sibling.n; k++ {
				child.K[k+nInit] = sibling.K[k]
				child.n++
			}
		}
	}
}

//func (node *Node[T]) fixChildUnderflow(i int, foundNode *Node[T], sibling *Node[T], rightSibling bool) {
//	if rightSibling {
//		foundNode.K[foundNode.n-1] = node.K[i]
//		if sibling.n > sibling.m-1 {
//			node.K[i] = sibling.K[0]
//			for K := 0; K < sibling.n-1; K++ {
//				sibling.K[K] = sibling.K[K+1]
//			}
//
//			keyDataType := reflect.TypeOf(sibling.K[sibling.n-1])
//			if keyDataType == reflect.TypeOf(0) {
//				sibling.K[sibling.n-1] = any(math.MinInt).(T) // Cast MinInt to type T
//			} else if keyDataType == reflect.TypeOf("") {
//				sibling.K[sibling.n-1] = any("").(T) // Cast "" to type T
//			}
//			sibling.n--
//		} else {
//			K := i
//			for K < node.n-1 {
//				node.K[K] = node.K[K+1]
//				node.C[K+1] = node.C[K+2]
//				K++
//			}
//			node.C[node.n] = nil
//
//			keyDataType := reflect.TypeOf(node.K[K])
//			if keyDataType == reflect.TypeOf(0) {
//				node.K[K] = any(math.MinInt).(T) // Cast MinInt to type T
//			} else if keyDataType == reflect.TypeOf("") {
//				node.K[K] = any("").(T) // Cast "" to type T
//			}
//			node.n--
//
//			nInit := foundNode.n
//			for K := 0; K < sibling.n; K++ {
//				foundNode.K[K+nInit] = sibling.K[K]
//				foundNode.n++
//			}
//		}
//	} else {
//		foundNode.K[0] = node.K[i-1]
//		if sibling.n > sibling.m-1 {
//			node.K[i-1] = sibling.K[sibling.n-1]
//
//			keyDataType := reflect.TypeOf(sibling.K[sibling.n-1])
//			if keyDataType == reflect.TypeOf(0) {
//				sibling.K[sibling.n-1] = any(math.MinInt).(T) // Cast MinInt to type T
//			} else if keyDataType == reflect.TypeOf("") {
//				sibling.K[sibling.n-1] = any("").(T) // Cast "" to type T
//			}
//			sibling.n--
//		} else {
//			for K := i - 1; K < node.n-1; K++ {
//				node.K[K] = node.K[K+1]
//				node.C[K] = node.C[K+1]
//			}
//			node.C[node.n-1] = node.C[node.n]
//			node.C[node.n] = nil
//
//			keyDataType := reflect.TypeOf(node.K[node.n-1])
//			if keyDataType == reflect.TypeOf(0) {
//				node.K[node.n-1] = any(math.MinInt).(T) // Cast MinInt to type T
//			} else if keyDataType == reflect.TypeOf("") {
//				node.K[node.n-1] = any("").(T) // Cast "" to type T
//			}
//			node.n--
//
//			nInit := foundNode.n
//			for K := sibling.n - 1; K >= 0; K-- {
//				foundNode.K[K+nInit] = foundNode.K[K]
//				foundNode.K[K] = sibling.K[K]
//				foundNode.n++
//			}
//		}
//	}
//}
