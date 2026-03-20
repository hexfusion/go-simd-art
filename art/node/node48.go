package node

const Node48Empty = 0xFF

// Node48 stores up to 48 children. Uses a 256-byte index array that maps
// key bytes to child slot positions, and a 48-element child array.
type Node48[V any] struct {
	Count    uint8
	Index    [256]byte
	Children [48]*Node[V]
}

func NewNode48[V any]() *Node48[V] {
	n := &Node48[V]{}
	for i := range n.Index {
		n.Index[i] = Node48Empty
	}
	return n
}

func (n *Node48[V]) FindChild(c byte) **Node[V] {
	idx := n.Index[c]
	if idx == Node48Empty {
		return nil
	}
	return &n.Children[idx]
}

func (n *Node48[V]) AddChild(c byte, child *Node[V]) {
	var slot byte
	for slot = 0; slot < 48; slot++ {
		if n.Children[slot] == nil {
			break
		}
	}
	n.Index[c] = slot
	n.Children[slot] = child
	n.Count++
}

func (n *Node48[V]) RemoveChild(c byte) {
	idx := n.Index[c]
	if idx == Node48Empty {
		return
	}
	n.Children[idx] = nil
	n.Index[c] = Node48Empty
	n.Count--
}
