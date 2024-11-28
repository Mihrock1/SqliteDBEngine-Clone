package storage

import "golang.org/x/exp/constraints"

type BTree[T constraints.Ordered] struct {
	root *Node[T]
	t    int
}

func NewBTree[T constraints.Ordered](t int) *BTree[T] {
	return &BTree[T]{
		root: NewNode[T](t, true),
		t:    t,
	}
}

func (btree *BTree[T]) insert(key T) {

}
