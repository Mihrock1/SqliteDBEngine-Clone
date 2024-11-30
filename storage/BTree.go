package storage

import (
	"errors"
	"golang.org/x/exp/constraints"
)

type BTree[T constraints.Ordered] struct {
	root *Node[T]
	t    int
}

func NewBTree[T constraints.Ordered](t int) (error, *BTree[T]) {
	if t < 2 {
		return errors.New("minimum degree must be greater than 2"), nil
	}

	return nil, &BTree[T]{
		root: newNode[T](t, true),
		t:    t,
	}
}

func (btree *BTree[T]) Insert(key T) {
	if btree.root.n == 2*btree.root.t-1 {
		oldRoot := btree.root
		btree.root = newNode[T](btree.t, false)
		btree.root.C[0] = oldRoot
		btree.root.splitChild(0, btree.root.C[0])
	}
	btree.root.insertNonFull(key)
}

func (btree *BTree[T]) search(key T) (error, *Node[T], int) {
	if btree.root.n == 0 {
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
	if btree.root.n == 0 {
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
	if btree.root.n == 0 {
		return errors.New("btree is empty"), false
	}

	err, foundNode, j, res := btree.root.deleteRec(key)
	if foundNode == btree.root {
		child := foundNode.C[0]
		if foundNode.leaf {
			err = foundNode.deleteFromLeaf(j)
		} else if foundNode.C[1] == nil && child.n < child.t {
			for k := j; k < foundNode.n-1; k++ {
				foundNode.keys[j] = foundNode.keys[j+1]
			}
			foundNode.n--

			nInit := foundNode.n
			for k := child.n - 1; k >= 0; k-- {
				foundNode.keys[k+nInit] = foundNode.keys[k]
				foundNode.keys[k] = child.keys[k]
				foundNode.n++
			}
			child = nil
		} else {
			var parent, C *Node[T]
			var val T
			if foundNode.C[j+1] != nil {
				parent, C, val = foundNode.findSmallestSubtreeKey(foundNode.C[j+1])
			} else if foundNode.C[j] != nil {
				parent, C, val = foundNode.findLargestSubtreeKey(foundNode.C[j])
			}
			foundNode.keys[j] = val
			if C.n > C.t-1 {
				err = C.deleteFromLeaf(0)
			} else {
				err = parent.deleteFromLeafWithTMinus1Keys(parent.n, C, C.n-1)
			}
		}
	}

	if err != nil {
		return err, res
	} else {
		return nil, true
	}
}

func (btree *BTree[T]) traverse() (error, []T) {
	if btree.root.n == 0 {
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
