package node

const (
	MaxPrefixLen = 10

	Node4Min   = 2
	Node4Max   = 4
	Node16Min  = 5
	Node16Max  = 16
	Node48Min  = 17
	Node48Max  = 48
	Node256Min = 49
)

// Kind identifies the type of an ART node.
type Kind byte

const (
	KindLeaf    Kind = iota
	KindNode4
	KindNode16
	KindNode48
	KindNode256
)

// Leaf holds a key-value pair at the leaves of the tree.
type Leaf[V any] struct {
	Key   []byte
	Value V
}

// Node is the internal representation of all ART node types.
type Node[V any] struct {
	Kind      Kind
	PrefixLen int
	Prefix    [MaxPrefixLen]byte

	Inner4   Node4[V]
	Inner16  Node16[V]
	Inner48  Node48[V]
	Inner256 Node256[V]
	LeafData Leaf[V]
}

func (n *Node[V]) IsLeaf() bool {
	return n.Kind == KindLeaf
}

func (n *Node[V]) Leaf() *Leaf[V] {
	return &n.LeafData
}

func NewLeaf[V any](key []byte, value V) *Node[V] {
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	return &Node[V]{
		Kind: KindLeaf,
		LeafData: Leaf[V]{
			Key:   keyCopy,
			Value: value,
		},
	}
}

// CheckPrefix compares the node's compressed prefix with the key at the given
// depth. Returns the number of matching bytes (up to PrefixLen).
func (n *Node[V]) CheckPrefix(key []byte, depth int) int {
	maxCmp := n.PrefixLen
	if maxCmp > MaxPrefixLen {
		maxCmp = MaxPrefixLen
	}
	if len(key)-depth < maxCmp {
		maxCmp = len(key) - depth
	}

	idx := 0
	for idx < maxCmp {
		if n.Prefix[idx] != key[depth+idx] {
			return idx
		}
		idx++
	}
	return idx
}

// NumChildren returns the number of children for this node.
func (n *Node[V]) NumChildren() int {
	switch n.Kind {
	case KindNode4:
		return int(n.Inner4.Count)
	case KindNode16:
		return int(n.Inner16.Count)
	case KindNode48:
		return int(n.Inner48.Count)
	case KindNode256:
		return int(n.Inner256.Count)
	default:
		return 0
	}
}

// FindChild returns a pointer to the child pointer for the given byte key,
// or nil if no such child exists.
func (n *Node[V]) FindChild(c byte) **Node[V] {
	switch n.Kind {
	case KindNode4:
		return n.Inner4.FindChild(c)
	case KindNode16:
		return n.Inner16.FindChild(c)
	case KindNode48:
		return n.Inner48.FindChild(c)
	case KindNode256:
		return n.Inner256.FindChild(c)
	default:
		return nil
	}
}

// AddChild adds a child to this node, growing the node type if necessary.
// Returns the (possibly new) node that replaces this one.
func (n *Node[V]) AddChild(c byte, child *Node[V]) *Node[V] {
	switch n.Kind {
	case KindNode4:
		if n.Inner4.Count < Node4Max {
			n.Inner4.AddChild(c, child)
			return n
		}
		return n.Grow4to16(c, child)
	case KindNode16:
		if n.Inner16.Count < Node16Max {
			n.Inner16.AddChild(c, child)
			return n
		}
		return n.Grow16to48(c, child)
	case KindNode48:
		if n.Inner48.Count < Node48Max {
			n.Inner48.AddChild(c, child)
			return n
		}
		return n.Grow48to256(c, child)
	case KindNode256:
		n.Inner256.AddChild(c, child)
		return n
	}
	return n
}

// RemoveChild removes a child, shrinking the node type if necessary.
// Returns the (possibly new) node that replaces this one.
func (n *Node[V]) RemoveChild(c byte) *Node[V] {
	switch n.Kind {
	case KindNode4:
		n.Inner4.RemoveChild(c)
		if n.Inner4.Count < 2 {
			return n.Shrink4()
		}
		return n
	case KindNode16:
		n.Inner16.RemoveChild(c)
		if n.Inner16.Count < Node16Min {
			return n.Shrink16to4()
		}
		return n
	case KindNode48:
		n.Inner48.RemoveChild(c)
		if n.Inner48.Count < Node48Min {
			return n.Shrink48to16()
		}
		return n
	case KindNode256:
		n.Inner256.RemoveChild(c)
		if n.Inner256.Count < Node256Min {
			return n.Shrink256to48()
		}
		return n
	}
	return n
}

// MatchKey compares two byte slices for equality.
func MatchKey(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// LongestCommonPrefix computes the length of the longest common prefix
// between two keys starting at the given depth.
func LongestCommonPrefix(a, b []byte, depth int) int {
	maxLen := len(a) - depth
	if len(b)-depth < maxLen {
		maxLen = len(b) - depth
	}
	idx := 0
	for idx < maxLen {
		if a[depth+idx] != b[depth+idx] {
			break
		}
		idx++
	}
	return idx
}

// Minimum returns the leftmost leaf in the subtree.
func Minimum[V any](n *Node[V]) *Leaf[V] {
	if n.IsLeaf() {
		return n.Leaf()
	}
	switch n.Kind {
	case KindNode4:
		return Minimum(n.Inner4.Children[0])
	case KindNode16:
		return Minimum(n.Inner16.Children[0])
	case KindNode48:
		for i := 0; i < 256; i++ {
			if n.Inner48.Index[i] != Node48Empty {
				return Minimum(n.Inner48.Children[n.Inner48.Index[i]])
			}
		}
	case KindNode256:
		for i := 0; i < 256; i++ {
			if n.Inner256.Children[i] != nil {
				return Minimum(n.Inner256.Children[i])
			}
		}
	}
	return nil
}

// Maximum returns the rightmost leaf in the subtree.
func Maximum[V any](n *Node[V]) *Leaf[V] {
	if n.IsLeaf() {
		return n.Leaf()
	}
	switch n.Kind {
	case KindNode4:
		return Maximum(n.Inner4.Children[n.Inner4.Count-1])
	case KindNode16:
		return Maximum(n.Inner16.Children[n.Inner16.Count-1])
	case KindNode48:
		for i := 255; i >= 0; i-- {
			if n.Inner48.Index[i] != Node48Empty {
				return Maximum(n.Inner48.Children[n.Inner48.Index[i]])
			}
		}
	case KindNode256:
		for i := 255; i >= 0; i-- {
			if n.Inner256.Children[i] != nil {
				return Maximum(n.Inner256.Children[i])
			}
		}
	}
	return nil
}
