package art

import "github.com/hexfusion/go-simd-art/art/node"

func (t *Tree[V]) insert(n *node.Node[V], ref **node.Node[V], key []byte, value V, depth int) (old V, replaced bool) {
	if n.IsLeaf() {
		leaf := n.Leaf()
		if node.MatchKey(leaf.Key, key) {
			old = leaf.Value
			leaf.Value = value
			return old, true
		}
		newNode := &node.Node[V]{Kind: node.KindNode4}
		lcp := node.LongestCommonPrefix(leaf.Key, key, depth)
		newNode.PrefixLen = lcp
		for i := 0; i < lcp && i < node.MaxPrefixLen; i++ {
			newNode.Prefix[i] = key[depth+i]
		}

		if depth+lcp < len(leaf.Key) {
			newNode.Inner4.AddChild(leaf.Key[depth+lcp], n)
		} else {
			newNode.Inner4.AddChild(0, n)
		}
		if depth+lcp < len(key) {
			newNode.Inner4.AddChild(key[depth+lcp], node.NewLeaf[V](key, value))
		} else {
			newNode.Inner4.AddChild(0, node.NewLeaf[V](key, value))
		}
		*ref = newNode
		var zero V
		return zero, false
	}

	if n.PrefixLen > 0 {
		mismatch := n.CheckPrefix(key, depth)
		if mismatch < n.PrefixLen {
			newNode := &node.Node[V]{Kind: node.KindNode4}
			newNode.PrefixLen = mismatch
			for i := 0; i < mismatch && i < node.MaxPrefixLen; i++ {
				newNode.Prefix[i] = n.Prefix[i]
			}

			if n.PrefixLen <= node.MaxPrefixLen {
				newNode.Inner4.AddChild(n.Prefix[mismatch], n)
				remaining := n.PrefixLen - mismatch - 1
				copy(n.Prefix[:], n.Prefix[mismatch+1:mismatch+1+remaining])
				n.PrefixLen = remaining
			} else {
				newNode.Inner4.AddChild(n.Prefix[mismatch], n)
				n.PrefixLen -= mismatch + 1
				copy(n.Prefix[:], n.Prefix[mismatch+1:])
			}

			if depth+mismatch < len(key) {
				newNode.Inner4.AddChild(key[depth+mismatch], node.NewLeaf[V](key, value))
			} else {
				newNode.Inner4.AddChild(0, node.NewLeaf[V](key, value))
			}
			*ref = newNode
			var zero V
			return zero, false
		}
		depth += n.PrefixLen
	}

	if depth >= len(key) {
		childRef := n.FindChild(0)
		if childRef != nil {
			return t.insert(*childRef, childRef, key, value, depth)
		}
		*ref = n.AddChild(0, node.NewLeaf[V](key, value))
		var zero V
		return zero, false
	}

	childRef := n.FindChild(key[depth])
	if childRef != nil {
		return t.insert(*childRef, childRef, key, value, depth+1)
	}

	*ref = n.AddChild(key[depth], node.NewLeaf[V](key, value))
	var zero V
	return zero, false
}

func (t *Tree[V]) delete(n *node.Node[V], ref **node.Node[V], key []byte, depth int) (value V, deleted bool) {
	if n == nil {
		var zero V
		return zero, false
	}

	if n.IsLeaf() {
		leaf := n.Leaf()
		if node.MatchKey(leaf.Key, key) {
			*ref = nil
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
		childRef := n.FindChild(0)
		if childRef == nil {
			var zero V
			return zero, false
		}
		leaf := (*childRef).Leaf()
		if node.MatchKey(leaf.Key, key) {
			value = leaf.Value
			*ref = n.RemoveChild(0)
			return value, true
		}
		var zero V
		return zero, false
	}

	childRef := n.FindChild(key[depth])
	if childRef == nil {
		var zero V
		return zero, false
	}

	value, deleted = t.delete(*childRef, childRef, key, depth+1)
	if deleted && *childRef == nil {
		*ref = n.RemoveChild(key[depth])
	}
	return value, deleted
}

func (t *Tree[V]) forEach(n *node.Node[V], fn func(key []byte, value V) bool) bool {
	if n.IsLeaf() {
		return fn(n.Leaf().Key, n.Leaf().Value)
	}

	switch n.Kind {
	case node.KindNode4:
		for i := 0; i < int(n.Inner4.Count); i++ {
			if !t.forEach(n.Inner4.Children[i], fn) {
				return false
			}
		}
	case node.KindNode16:
		for i := 0; i < int(n.Inner16.Count); i++ {
			if !t.forEach(n.Inner16.Children[i], fn) {
				return false
			}
		}
	case node.KindNode48:
		for i := 0; i < 256; i++ {
			if n.Inner48.Index[i] != node.Node48Empty {
				if !t.forEach(n.Inner48.Children[n.Inner48.Index[i]], fn) {
					return false
				}
			}
		}
	case node.KindNode256:
		for i := 0; i < 256; i++ {
			if n.Inner256.Children[i] != nil {
				if !t.forEach(n.Inner256.Children[i], fn) {
					return false
				}
			}
		}
	}
	return true
}

func (t *Tree[V]) forEachPrefix(n *node.Node[V], prefix []byte, depth int, fn func(key []byte, value V) bool) bool {
	if n == nil {
		return true
	}

	if n.IsLeaf() {
		leaf := n.Leaf()
		if len(leaf.Key) >= len(prefix) && node.MatchKey(leaf.Key[:len(prefix)], prefix) {
			return fn(leaf.Key, leaf.Value)
		}
		return true
	}

	if n.PrefixLen > 0 {
		maxCmp := n.PrefixLen
		if maxCmp > node.MaxPrefixLen {
			maxCmp = node.MaxPrefixLen
		}
		for i := 0; i < maxCmp && depth+i < len(prefix); i++ {
			if n.Prefix[i] != prefix[depth+i] {
				return true
			}
		}
		depth += n.PrefixLen
	}

	if depth >= len(prefix) {
		return t.forEach(n, fn)
	}

	childRef := n.FindChild(prefix[depth])
	if childRef != nil {
		return t.forEachPrefix(*childRef, prefix, depth+1, fn)
	}
	return true
}

func (t *Tree[V]) longestPrefix(n *node.Node[V], key []byte, depth int) (matchedKey []byte, value V, found bool) {
	if n == nil {
		var zero V
		return nil, zero, false
	}

	if n.IsLeaf() {
		leaf := n.Leaf()
		if len(key) >= len(leaf.Key) && node.MatchKey(key[:len(leaf.Key)], leaf.Key) {
			return leaf.Key, leaf.Value, true
		}
		var zero V
		return nil, zero, false
	}

	if n.PrefixLen > 0 {
		mismatch := n.CheckPrefix(key, depth)
		if mismatch != n.PrefixLen {
			var zero V
			return nil, zero, false
		}
		depth += n.PrefixLen
	}

	childRef := n.FindChild(0)
	if childRef != nil && (*childRef).IsLeaf() {
		leaf := (*childRef).Leaf()
		if len(key) >= len(leaf.Key) && node.MatchKey(key[:len(leaf.Key)], leaf.Key) {
			matchedKey, value, found = leaf.Key, leaf.Value, true
		}
	}

	if depth < len(key) {
		nextRef := n.FindChild(key[depth])
		if nextRef != nil {
			deepKey, deepVal, deepFound := t.longestPrefix(*nextRef, key, depth+1)
			if deepFound {
				return deepKey, deepVal, true
			}
		}
	}

	return matchedKey, value, found
}
