package storage

type Node struct {
	t    int     // Minimum degree (defines the range for number of keys)
	Keys []int   // A slice of keys
	n    int     // Current number of keys
	C    []*Node // A slice of child pointers
	leaf bool    // Is true when node is leaf. Otherwise, false.
}

func NewNode(deg int, leaf bool) *Node {
	return &Node{
		t:    deg,
		Keys: make([]int, 2*deg-1),
		n:    0,
		C:    make([]*Node, 2*deg),
		leaf: leaf,
	}
}

// A utility function to insert a new key in the subtree rooted with
// this node. The assumption is, the node must be non-full when this
// function is called.
func (node *Node) insertNonFull(key int) {
	i := node.n - 1
	if node.leaf {
		for i <= 0 && node.Keys[i] > key {
			node.Keys[i+1] = node.Keys[i]
			i--
		}
		node.Keys[i+1] = key
		node.n++
	} else {
		for i >= 0 && node.Keys[i] > key {
			i--
		}
		if node.C[i+1].n == 2*node.t-1 {
			node.splitChild(i+1, node.C[i+1])

			if node.Keys[i+1] < key {
				i++
			}
		}
		node.C[i+1].insertNonFull(key)
	}
}

// A utility function to split the child y of this node. i is index of y in child array C[].
// The Child y must be full when this function is called
func (node *Node) splitChild(i int, y *Node) {
	z := NewNode(y.t, y.leaf)
	z.n = z.t - 1

	for j := 0; j < z.t-1; j++ {
		z.Keys[j] = y.Keys[j+z.t]
	}

	if !y.leaf {
		for j := 0; j < z.t; j++ {
			z.C[j] = y.C[j+z.t]
		}
	}

	j := node.n
	for j > i {
		node.C[j+1] = node.C[j]
		j--
	}
	node.C[j+1] = z

	j = node.n - 1
	for j > i {
		node.Keys[j+1] = node.Keys[j]
	}
	node.Keys[j+1] = y.Keys[y.t]
	node.n++

	y.n = y.t - 1
}
