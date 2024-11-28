package storage

import "golang.org/x/exp/constraints"

// T is constrained to ordered types
type T interface {
	constraints.Ordered
}

type Node[T constraints.Ordered] struct {
	t    int        // Minimum degree (defines the range for number of keys)
	keys []T        // A slice of keys
	n    int        // Current number of keys
	C    []*Node[T] // A slice of child pointers
	leaf bool       // Is true when node is leaf. Otherwise, false
}

func NewNode[T constraints.Ordered](t int, leaf bool) *Node[T] {
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
	z := NewNode[T](y.t, y.leaf)
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
