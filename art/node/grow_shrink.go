package node

func (n *Node[V]) Grow4to16(c byte, child *Node[V]) *Node[V] {
	n.Kind = KindNode16
	n.Inner16 = Node16[V]{}
	for i := 0; i < int(n.Inner4.Count); i++ {
		n.Inner16.AddChild(n.Inner4.Keys[i], n.Inner4.Children[i])
	}
	n.Inner4 = Node4[V]{}
	n.Inner16.AddChild(c, child)
	return n
}

func (n *Node[V]) Grow16to48(c byte, child *Node[V]) *Node[V] {
	n48 := NewNode48[V]()
	for i := 0; i < int(n.Inner16.Count); i++ {
		n48.AddChild(n.Inner16.Keys[i], n.Inner16.Children[i])
	}
	n48.AddChild(c, child)

	n.Kind = KindNode48
	n.Inner16 = Node16[V]{}
	n.Inner48 = *n48
	return n
}

func (n *Node[V]) Grow48to256(c byte, child *Node[V]) *Node[V] {
	n.Kind = KindNode256
	n.Inner256 = Node256[V]{}
	for i := 0; i < 256; i++ {
		if n.Inner48.Index[i] != Node48Empty {
			n.Inner256.AddChild(byte(i), n.Inner48.Children[n.Inner48.Index[i]])
		}
	}
	n.Inner48 = Node48[V]{}
	n.Inner256.AddChild(c, child)
	return n
}

func (n *Node[V]) Shrink4() *Node[V] {
	child := n.Inner4.Children[0]
	if child.IsLeaf() {
		return child
	}
	prefix := make([]byte, 0, n.PrefixLen+1+child.PrefixLen)
	for i := 0; i < n.PrefixLen && i < MaxPrefixLen; i++ {
		prefix = append(prefix, n.Prefix[i])
	}
	prefix = append(prefix, n.Inner4.Keys[0])
	for i := 0; i < child.PrefixLen && i < MaxPrefixLen; i++ {
		prefix = append(prefix, child.Prefix[i])
	}
	child.PrefixLen = len(prefix)
	copy(child.Prefix[:], prefix)
	return child
}

func (n *Node[V]) Shrink16to4() *Node[V] {
	n.Kind = KindNode4
	n.Inner4 = Node4[V]{}
	for i := 0; i < int(n.Inner16.Count); i++ {
		n.Inner4.AddChild(n.Inner16.Keys[i], n.Inner16.Children[i])
	}
	n.Inner16 = Node16[V]{}
	return n
}

func (n *Node[V]) Shrink48to16() *Node[V] {
	n.Kind = KindNode16
	n.Inner16 = Node16[V]{}
	for i := 0; i < 256; i++ {
		if n.Inner48.Index[i] != Node48Empty {
			n.Inner16.AddChild(byte(i), n.Inner48.Children[n.Inner48.Index[i]])
		}
	}
	n.Inner48 = Node48[V]{}
	return n
}

func (n *Node[V]) Shrink256to48() *Node[V] {
	n48 := NewNode48[V]()
	for i := 0; i < 256; i++ {
		if n.Inner256.Children[i] != nil {
			n48.AddChild(byte(i), n.Inner256.Children[i])
		}
	}
	n.Kind = KindNode48
	n.Inner256 = Node256[V]{}
	n.Inner48 = *n48
	return n
}
