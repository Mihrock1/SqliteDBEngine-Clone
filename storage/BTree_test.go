package storage

import (
	"testing"
)

func TestSearch(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Test empty tree
	if btree.Exists(10) {
		t.Error("Expected false for empty tree search")
	}

	// Insert some K
	keys := []int{10, 20, 30, 40, 50}
	for _, key := range keys {
		btree.Insert(key)
	}

	// Test existing K
	for _, key := range keys {
		if !btree.Exists(key) {
			t.Errorf("Expected to find key %d", key)
		}
	}

	// Test non-existing K
	nonExistingKeys := []int{15, 25, 35, 45}
	for _, key := range nonExistingKeys {
		if btree.Exists(key) {
			t.Errorf("Did not expect to find key %d", key)
		}
	}
}

func TestTraversal(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Test empty tree
	err, keys := btree.traverse()
	if err == nil {
		t.Error("Expected error for empty tree traversal")
	}
	if keys != nil {
		t.Error("Expected nil K for empty tree")
	}

	// Insert K in random m
	insertKeys := []int{50, 30, 10, 40, 20}
	for _, key := range insertKeys {
		btree.Insert(key)
	}

	// Check if traversal returns sorted K
	err, keys = btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Verify K are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted m: %v", keys)
			break
		}
	}
}

func TestStringKeys(t *testing.T) {
	deg := 3
	_, btree := NewBTree[string](deg)

	// Test string K
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

	// Verify traversal m
	err, keys := btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during string traversal: %v", err)
	}

	// Check if strings are in lexicographical m
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Strings not in lexicographical m: %v", keys)
			break
		}
	}
}

func TestLargeNumberOfKeys(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert 100 K
	for i := 0; i < 100; i++ {
		btree.Insert(i)
	}

	// Verify all K exist
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

	// Check if K are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted m: %v", keys)
			break
		}
	}
}

func TestMinimumDegreeValidation(t *testing.T) {
	// Test degrees less than 2
	err, _ := NewBTree[int](1)
	if err == nil {
		t.Error("Expected panic for minimum degree < 2")
	}
}

func TestNodeFullness(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert 2t-1 K to fill root
	for i := 0; i < deg-1; i++ {
		btree.Insert(i)
	}

	if btree.root.n != deg-1 {
		t.Errorf("Expected root to have %d K, got %d", deg-1, btree.root.n)
	}

	// Insert one more key to force split
	btree.Insert(deg - 1)

	if btree.root.n >= deg-1 {
		t.Error("Expected root to split")
	}
}

func TestDeleteFromEmptyTree(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	err, success := btree.Delete(10)
	if err == nil {
		t.Error("Expected error when deleting from empty tree")
	}
	if success {
		t.Error("Expected delete operation to fail on empty tree")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert some K
	for i := 1; i <= 5; i++ {
		btree.Insert(i * 10)
	}

	// Try to delete a non-existent key
	err, success := btree.Delete(15)
	if err == nil {
		t.Error("Expected error when deleting non-existent key")
	}
	if success {
		t.Error("Expected delete operation to fail for non-existent key")
	}
}

func TestDeleteFromLeaf(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert K
	keys := []int{10, 20, 30, 40, 50}
	for _, key := range keys {
		btree.Insert(key)
	}

	// Delete a isLeaf node key
	err, success := btree.Delete(50)
	if err != nil {
		t.Errorf("Unexpected error during deletion: %v", err)
	}
	if !success {
		t.Error("Expected delete operation to succeed")
	}

	// Verify key no longer exists
	if btree.Exists(50) {
		t.Error("Key should not exist after deletion")
	}

	// Verify remaining K are intact
	for _, key := range []int{10, 20, 30, 40} {
		if !btree.Exists(key) {
			t.Errorf("Key %d should still exist after deletion", key)
		}
	}
}

func TestDeleteFromInternalNode(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert enough K to create internal nodes
	for i := 1; i <= 10; i++ {
		btree.Insert(i * 10)
	}

	// Delete a key that should be in an internal node
	err, success := btree.Delete(50)
	if err != nil {
		t.Errorf("Unexpected error during deletion: %v", err)
	}
	if !success {
		t.Error("Expected delete operation to succeed")
	}

	// Verify the tree structure remains valid
	err, keys := btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Check if remaining K are sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted m after deletion: %v", keys)
		}
	}
}

func TestDeleteWithKeyBorrowing(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert K to create a scenario where borrowing will be needed
	keys := []int{10, 20, 30, 40, 50, 60, 70}
	for _, key := range keys {
		btree.Insert(key)
	}

	// Delete K that will trigger borrowing
	err, success := btree.Delete(30)
	if err != nil {
		t.Errorf("Unexpected error during deletion: %v", err)
	}
	if !success {
		t.Error("Expected delete operation to succeed")
	}

	// Verify tree properties after borrowing
	err, keys = btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Verify K remain sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Errorf("Keys not in sorted m after borrowing: %v", keys)
		}
	}
}

func TestDeleteWithNodeMerging(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert K to create a scenario where merging will be needed
	keys := []int{10, 20, 30, 40, 50, 60}
	for _, key := range keys {
		btree.Insert(key)
	}

	// Delete multiple K to force node merging
	deleteKeys := []int{20, 40, 60}
	for _, key := range deleteKeys {
		err, success := btree.Delete(key)
		if err != nil {
			t.Errorf("Unexpected error during deletion: %v", err)
		}
		if !success {
			t.Error("Expected delete operation to succeed")
		}
	}

	// Verify remaining K
	err, remainingKeys := btree.traverse()
	if err != nil {
		t.Errorf("Unexpected error during traversal: %v", err)
	}

	// Check if remaining K are sorted
	for i := 1; i < len(remainingKeys); i++ {
		if remainingKeys[i] <= remainingKeys[i-1] {
			t.Errorf("Keys not in sorted m after merging: %v", remainingKeys)
		}
	}
}

func TestSequentialDeletion(t *testing.T) {
	deg := 3
	_, btree := NewBTree[int](deg)

	// Insert K
	for i := 1; i <= 20; i++ {
		btree.Insert(i)
	}

	// Delete all K sequentially
	for i := 1; i <= 20; i++ {
		err, success := btree.Delete(i)
		if err != nil {
			t.Errorf("Unexpected error deleting key %d: %v", i, err)
		}
		if !success {
			t.Errorf("Failed to delete key %d", i)
		}

		// Verify deleted key no longer exists
		if btree.Exists(i) {
			t.Errorf("Key %d still exists after deletion", i)
		}
	}

	// Verify tree is empty
	err, keys := btree.traverse()
	if err == nil {
		t.Error("Expected error when traversing empty tree")
	}
	if keys != nil {
		t.Error("Expected nil K for empty tree")
	}
}
