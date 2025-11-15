package storage

import (
	"errors"

	"golang.org/x/exp/constraints"
)

/*
“We define the minimum degree t of a B-tree as the minimum number of children any node can have.
Each node contains between t - 1 & 2*t - 1 keys, and between t and 2t children.”

Each node can have at most 2t children (so, at most 2t – 1 keys).

Each internal node (except root) has at least t children (so, at least t – 1 keys).

TODO: Recheck below calculation
For each node or Page:
TODO: check if 8 bytes is really needed considering all datatypes
Needs space for up to 2*t-1 keys (8 bytes each)

Needs space for up to 2*t child node pointers (8 bytes each)

Plus some fixed metadata (say 16 bytes)

Total max size: 8*(2*t - 1) + 8*(2*t) + 16 = 16*t - 8 + 16*t + 16 = 32t + 8

So, minimum degree t of BTree can be derived from equation: 32t + 8 <= page size of disk
*/

type BTree[T constraints.Ordered] struct {
	root *Node[T]
	t    int
	h    int
}

func NewBTree[T constraints.Ordered](minDegree int) (error, *BTree[T]) {
	// TODO: check if below check is needed
	if minDegree < 2 {
		return errors.New("minimum degree must be greater than 2"), nil
	}

	return nil, &BTree[T]{
		root: nil,
		t:    minDegree,
		h:    0,
	}
}

// TODO: To implement later
//func NewBTree[T constraints.Ordered](pageSize int) (error, *BTree[T]) {
//	t := (pageSize - 8) / 32
//
//	return nil, &BTree[T]{
//		root:   nil,
//		t:      t,
//		h: 0,
//	}
//}

func (btree *BTree[T]) Insert(key T) {
	if btree.root == nil {
		btree.root = newNode[T](btree.t, true)
		btree.root.K[0] = key
		btree.root.n = 1

	} else if btree.root.n == 2*btree.root.t-1 {
		oldRoot := btree.root
		btree.root = newNode[T](btree.t, false)
		btree.root.C[0] = oldRoot
		btree.root.splitChild(0)

		idx := 0
		if btree.root.K[0] < key {
			idx = 1
		}
		btree.root.C[idx].insertNonFull(key)

	} else {
		btree.root.insertNonFull(key)
	}
}

func (btree *BTree[T]) search(key T) (error, *Node[T], int) {
	if btree.root == nil || (btree.root.n == 0 && btree.root.leaf) {
		//defaultVal := defaultValue[T]()
		return errors.New("the btree is empty"), nil, 0
	}

	err, node, i := btree.root.searchRec(key)
	if err != nil {
		return err, nil, 0
	} else {
		return nil, node, i
	}
}

func (btree *BTree[T]) Exists(key T) bool {
	// TODO: test new logic below
	if btree.root == nil || btree.root.n == 0 {
		return false
	}

	err, _, _ := btree.search(key)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (btree *BTree[T]) Delete(key T) (error, bool) {
	if btree.root == nil || (btree.root.n == 0 && btree.root.leaf) {
		return errors.New("btree is empty"), false
	}
	//else if btree.root.n == 0 && !btree.root.leaf {
	//	btree.root = btree.root.C[0]
	//}

	err, foundNode, j, _ := btree.root.deleteRec(key)
	if err != nil {
		return err, false
	}

	if foundNode == btree.root {
		firstChild := btree.root.C[0]
		secondChild := btree.root.C[1]
		if btree.root.leaf {
			btree.root.deleteFromLeaf(j)
		} else if !btree.root.leaf && btree.root.n == 1 {
			if secondChild == nil {
				btree.root = firstChild
			} else {
				nInit := firstChild.n - 1
				for k := 0; k < secondChild.n; k++ {
					firstChild.K[nInit+k] = secondChild.K[k]
				}
				btree.root = firstChild
			}
		} else {
			//var searchIn, P, C *Node[T]
			//var val T
			//var atIndex int
			//if btree.root.C[j] != nil {
			//	atIndex = btree.root.C[j].n - 1
			//	searchIn = btree.root.C[j]
			//} else if btree.root.C[j+1] != nil {
			//	atIndex = 0
			//	searchIn = btree.root.C[j+1]
			//}
			//P, C, val = btree.root.findKeyInSubtreeRec(searchIn, atIndex)
			//btree.root.K[j] = val
			//if C.n > C.t-1 {
			//	C.deleteFromLeaf(atIndex)
			//} else {
			//	P.deleteFromLeafWithTMinus1Keys(atIndex, C, atIndex)
			//}
		}
	}
	return nil, true
}

func (btree *BTree[T]) traverse() (error, []T) {
	if btree.root == nil || (btree.root.n == 0 && btree.root.leaf) {
		return errors.New("the btree is empty"), nil
	}
	var keys []T
	btree.root.traverseRec(keys)
	return nil, keys
}

func (btree *BTree[T]) Print() {
	err, keys := btree.traverse()
	if err == nil {
		println(err)
	} else {
		for key := range keys {
			print(key, " ")
		}
		println()
	}
}
