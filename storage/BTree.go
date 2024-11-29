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
