package art

import "github.com/hexfusion/go-simd-art/art/node"

// Tree is an Adaptive Radix Tree.
type Tree[V any] struct {
	root *node.Node[V]
	size int
}

// New creates an empty ART.
func New[V any]() *Tree[V] {
	return &Tree[V]{}
}

// Size returns the number of key-value pairs in the tree.
func (t *Tree[V]) Size() int {
	return t.size
}

// Insert adds or updates a key-value pair. Returns the previous value and
// whether the key already existed.
func (t *Tree[V]) Insert(key []byte, value V) (old V, replaced bool) {
	if t.root == nil {
		t.root = node.NewLeaf[V](key, value)
		t.size++
		var zero V
		return zero, false
	}
	old, replaced = t.insert(t.root, &t.root, key, value, 0)
	if !replaced {
		t.size++
	}
	return old, replaced
}

// Search returns the value for the given key and whether it was found.
func (t *Tree[V]) Search(key []byte) (value V, found bool) {
	n := t.root
	depth := 0

	for n != nil {
		if n.IsLeaf() {
			leaf := n.Leaf()
			if node.MatchKey(leaf.Key, key) {
				return leaf.Value, true
			}
			var zero V
			return zero, false
		}

		if n.PrefixLen > 0 {
			mismatch := n.CheckPrefix(key, depth)
			if mismatch != n.PrefixLen {
				var zero V
				return zero, false
			}
			depth += n.PrefixLen
		}

		if depth >= len(key) {
			next := n.FindChild(0)
			if next == nil {
				var zero V
				return zero, false
			}
			n = *next
			continue
		}

		next := n.FindChild(key[depth])
		if next == nil {
			var zero V
			return zero, false
		}
		n = *next
		depth++
	}
	var zero V
	return zero, false
}

// Delete removes a key and returns its value. Returns false if key not found.
func (t *Tree[V]) Delete(key []byte) (value V, deleted bool) {
	if t.root == nil {
		var zero V
		return zero, false
	}
	value, deleted = t.delete(t.root, &t.root, key, 0)
	if deleted {
		t.size--
	}
	return value, deleted
}

// ForEach calls fn for every key-value pair in sorted order.
// If fn returns false, iteration stops.
func (t *Tree[V]) ForEach(fn func(key []byte, value V) bool) {
	if t.root != nil {
		t.forEach(t.root, fn)
	}
}

// ForEachPrefix calls fn for every key-value pair whose key starts with
// the given prefix, in sorted order. If fn returns false, iteration stops.
func (t *Tree[V]) ForEachPrefix(prefix []byte, fn func(key []byte, value V) bool) {
	if t.root == nil {
		return
	}
	t.forEachPrefix(t.root, prefix, 0, fn)
}

// LongestPrefix returns the key-value pair with the longest key that is
// a prefix of the given key. Returns false if no prefix match exists.
func (t *Tree[V]) LongestPrefix(key []byte) (matchKey []byte, value V, found bool) {
	if t.root == nil {
		var zero V
		return nil, zero, false
	}
	return t.longestPrefix(t.root, key, 0)
}

// Minimum returns the smallest key-value pair. Returns false if empty.
func (t *Tree[V]) Minimum() (key []byte, value V, found bool) {
	if t.root == nil {
		var zero V
		return nil, zero, false
	}
	leaf := node.Minimum(t.root)
	return leaf.Key, leaf.Value, true
}

// Maximum returns the largest key-value pair. Returns false if empty.
func (t *Tree[V]) Maximum() (key []byte, value V, found bool) {
	if t.root == nil {
		var zero V
		return nil, zero, false
	}
	leaf := node.Maximum(t.root)
	return leaf.Key, leaf.Value, true
}
