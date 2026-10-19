# Repository agent instructions

## Project overview

This is a small Go 1.22+ learning project for a SQLite-inspired storage engine. The implemented code is currently concentrated in the `storage` package: `BTree[T]` owns tree-level operations, while `Node[T]` implements recursive insertion, search, traversal, and deletion. There is currently no executable, SQL parser, persistence layer, or separate query-processing package.

The B-tree accepts `constraints.Ordered` keys and requires a minimum degree `t >= 2`. Nodes allocate `2*t-1` key slots and `2*t` child slots. Insertion splits full children before descending. Deletion removes leaf keys or replaces internal keys from a subtree, then repairs underflow through sibling borrowing or node merging.

## Commands

Run from the repository root:

```text
go test ./...
go test ./storage -run '^TestSearch$'
go vet ./...
gofmt -w storage\BTree.go storage\BTreeNode.go storage\BTree_test.go
```

Use `go test ./storage -run '^TestName$'` for a single test or focused test pattern. Tests are in `storage/BTree_test.go` and use the `storage` package, allowing them to exercise unexported helpers such as `traverse`.

## Repository conventions

- Keep public tree APIs generic over `constraints.Ordered`; comparison and ascending traversal are core behavior.
- Keep tree-level orchestration in `BTree.go` and recursive node mechanics in `BTreeNode.go`. Reuse existing helpers such as `insertNonFull`, `splitChild`, `searchRec`, `deleteRec`, and `fixUnderflow`.
- Constructors return `(error, *BTree[T])`; minimum degrees below 2 are invalid. Empty-tree and missing-key operations report errors and failure results.
- Maintain `n` whenever keys move between nodes, and keep child pointers aligned with separator keys during splits, borrowing, and merging.
- When changing structural algorithms, test leaf deletion, internal-node deletion, borrowing, merging, empty trees, integer keys, string keys, and sorted traversal.
- Preserve the existing Go formatting, error-return style, and package layout. Keep experimental alternatives consistent with the surrounding TODOs instead of adding a parallel implementation.
