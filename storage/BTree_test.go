package storage

import (
	"testing"
)

func TestSearch(t *testing.T) {
	deg := 3
	btree := NewBTree[int](deg)

	// Test empty tree
	if btree.Exists(10) {
		t.Error("Expected false for empty tree search")
	}

	// Insert some keys
	keys := []int{10, 20, 30, 40, 50}
	for _, key := range keys {
		btree.Insert(key)
	}

	// Test existing keys
	for _, key := range keys {
		if !btree.Exists(key) {
			t.Errorf("Expected to find key %d", key)
		}
	}

	// Test non-existing keys
	nonExistingKeys := []int{15, 25, 35, 45}
	for _, key := range nonExistingKeys {
		if btree.Exists(key) {
			t.Errorf("Did not expect to find key %d", key)
		}
	}
}

func TestTraversal(t *testing.T) {
	deg := 3
	btree := NewBTree[int](deg)

	// Test empty tree
	err, keys := btree.traverse()
	if err == nil {
		t.Error("Expected error for empty tree traversal")
	}
	if keys != nil {
		t.Error("Expected nil keys for empty tree")
	}

	// Insert keys in random order
	insertKeys := []int{50, 30, 10, 40, 20}
	for _, key := range insertKeys {
		btree.Insert(key)
	}

	// Check if traversal returns sorted keys
	err, keys = btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Verify keys are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted order: %v", keys)
			break
		}
	}
}

func TestStringKeys(t *testing.T) {
	deg := 3
	btree := NewBTree[string](deg)

	// Test string keys
	strings := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for _, s := range strings {
		btree.Insert(s)
	}

	// Verify all strings exist
	for _, s := range strings {
		if !btree.Exists(s) {
			t.Errorf("Expected to find string %s", s)
		}
	}

	// Verify traversal order
	err, keys := btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during string traversal: %v", err)
	}

	// Check if strings are in lexicographical order
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Strings not in lexicographical order: %v", keys)
			break
		}
	}
}

func TestLargeNumberOfKeys(t *testing.T) {
	deg := 3
	btree := NewBTree[int](deg)

	// Insert 100 keys
	for i := 0; i < 100; i++ {
		btree.Insert(i)
	}

	// Verify all keys exist
	for i := 0; i < 100; i++ {
		if !btree.Exists(i) {
			t.Errorf("Expected to find key %d", i)
		}
	}

	// Verify tree properties
	err, keys := btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Check if keys are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted order: %v", keys)
			break
		}
	}
}

func TestMinimumDegreeValidation(t *testing.T) {
	// Test degrees less than 2
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for minimum degree < 2")
		}
	}()

	NewBTree[int](1) // Should panic
}

func TestNodeFullness(t *testing.T) {
	deg := 3
	btree := NewBTree[int](deg)

	// Insert 2t-1 keys to fill root
	for i := 0; i < 2*deg-1; i++ {
		btree.Insert(i)
	}

	if btree.root.n != 2*deg-1 {
		t.Errorf("Expected root to have %d keys, got %d", 2*deg-1, btree.root.n)
	}

	// Insert one more key to force split
	btree.Insert(2*deg - 1)

	if btree.root.n >= 2*deg-1 {
		t.Error("Expected root to split")
	}
}
