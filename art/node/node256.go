package node

// Node256 stores up to 256 children, directly indexed by key byte.
type Node256[V any] struct {
	Count    uint16
	Children [256]*Node[V]
}

func (n *Node256[V]) FindChild(c byte) **Node[V] {
	if n.Children[c] == nil {
		return nil
	}
	return &n.Children[c]
}

func (n *Node256[V]) AddChild(c byte, child *Node[V]) {
	n.Children[c] = child
	n.Count++
}

func (n *Node256[V]) RemoveChild(c byte) {
	if n.Children[c] != nil {
		n.Children[c] = nil
		n.Count--
	}
}
