package storage

import (
	"errors"
	"reflect"

	"golang.org/x/exp/constraints"
)

//func defaultValue[T any]() T {
//	var defaultVal T
//	return defaultVal
//}

// TODO: Can we replace slices with arrays of max keys ?
// TODO: check if t is needed here when it already exists in btree
type Node[T constraints.Ordered] struct {
	t    int        // min Degree of BTree Node
	n    int        // Current number of keys
	K    []T        // A slice of keys
	C    []*Node[T] // A slice of child pointers
	leaf bool       // Is true when node is leaf. Otherwise, false
}

func newNode[T constraints.Ordered](minDegree int, leaf bool) *Node[T] {
	return &Node[T]{
		t:    minDegree,
		K:    make([]T, 2*minDegree-1),
		n:    0,
		C:    make([]*Node[T], 2*minDegree),
		leaf: leaf,
	}
}

func (node *Node[T]) insertNonFull(key T) {
	i := node.n - 1
	if node.leaf {
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

		// bring index value back to condition node.K[i] > key
		i++

		// in case child node is full
		if node.C[i].n == 2*node.t-1 {
			//node.splitChild(i + 1)
			node.splitChild(i)

			// TODO: Check if new logic below works
			// after split, fix index position if condition not met: node.K[i] > key
			if node.K[i] < key {
				i++
			}
		}

		node.C[i].insertNonFull(key)
	}
}

// TODO: check if two params are needed
func (node *Node[T]) splitChild(i int) {
	child := node.C[i]
	newChild := newNode[T](child.t, child.leaf)
	// TODO: update value of n as key in each node is added or deleted to make debugging easier
	//newChild.n = child.t - 1

	var zero T

	for j := 0; j < child.t-1; j++ {
		// TODO: Check for problems with debugging here
		// move keys from second half of first child to new child
		newChild.K[j] = child.K[j+child.t]
		newChild.n++
		child.K[j+child.t] = zero
		child.n--

		// move child pointers from second half of first child to new child
		newChild.C[j] = child.C[j+child.t]
		child.C[j+child.t] = nil
	}

	// for last child pointer not encountered in above loop
	newChild.C[child.t-1] = child.C[2*child.t-1]
	child.C[2*child.t-1] = nil

	//if !child.leaf {
	//	for j := 0; j < child.t; j++ {
	//		newChild.C[j] = child.C[j+child.t]
	//		child.C[j+child.t] = nil
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
	// moving last key on child to parent at index i
	node.K[i] = child.K[child.t-1]
	node.n++
	child.K[child.t-1] = zero
	child.n--

	//child.n = child.t - 1
}

func (node *Node[T]) traverseRec(keys []T) {
	for i := 0; i < node.n; i++ {
		if !node.leaf {
			node.C[i].traverseRec(keys)
		}
		keys = append(keys, node.K[i])
	}
	if !node.leaf {
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
	if node.leaf {
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
	if node.leaf {
		return errors.New("key does not exist in btree"), nil, -1, false
	} else {
		err, child, j, res := node.C[i].deleteRec(key)
		if err != nil {
			return err, child, j, res
		} else if res == true {
			if child.n < child.t-1 {
				node.fixUnderflow(i, child)
			}
			return nil, node, i, true
		} else {
			if child.leaf {
				if child.n > child.t-1 {
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
				if C.n > C.t-1 {
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

		if sibling.n > sibling.t-1 {
			node.K[i-1] = sibling.K[sibling.n-1]

			P, C, val := child.findLargestKeyInSubtreeRec(sibling.C[sibling.n-1])
			sibling.K[sibling.n-1] = val
			if C.n > C.t-1 {
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

		if sibling.n > sibling.t-1 {
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
//		if sibling.n > sibling.t-1 {
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
//		if sibling.n > sibling.t-1 {
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
