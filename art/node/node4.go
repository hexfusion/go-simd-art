package node

// Node4 stores up to 4 children with sorted keys.
type Node4[V any] struct {
	Count    uint8
	Keys     [4]byte
	Children [4]*Node[V]
}

func (n *Node4[V]) FindChild(c byte) **Node[V] {
	for i := 0; i < int(n.Count); i++ {
		if n.Keys[i] == c {
			return &n.Children[i]
		}
	}
	return nil
}

func (n *Node4[V]) AddChild(c byte, child *Node[V]) {
	idx := int(n.Count)
	for i := 0; i < int(n.Count); i++ {
		if c < n.Keys[i] {
			idx = i
			break
		}
	}
	for i := int(n.Count); i > idx; i-- {
		n.Keys[i] = n.Keys[i-1]
		n.Children[i] = n.Children[i-1]
	}
	n.Keys[idx] = c
	n.Children[idx] = child
	n.Count++
}

func (n *Node4[V]) RemoveChild(c byte) {
	for i := 0; i < int(n.Count); i++ {
		if n.Keys[i] == c {
			for j := i; j < int(n.Count)-1; j++ {
				n.Keys[j] = n.Keys[j+1]
				n.Children[j] = n.Children[j+1]
			}
			n.Count--
			n.Keys[n.Count] = 0
			n.Children[n.Count] = nil
			return
		}
	}
}
