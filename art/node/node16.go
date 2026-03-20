package node

import (
	"math/bits"
	"simd/archsimd"
	"unsafe"
)

// Node16 stores up to 16 children with sorted keys.
// FindChild uses SSE SIMD to compare all 16 keys in a single instruction.
type Node16[V any] struct {
	Count    uint8
	Keys     [16]byte
	Children [16]*Node[V]
}

// FindChild uses SIMD to compare the search byte against all 16 keys
// simultaneously. This is the key optimization from the Leis paper:
// broadcast the search byte to all 16 lanes, compare in one instruction,
// extract match positions from the resulting bitmask.
func (n *Node16[V]) FindChild(c byte) **Node[V] {
	// Load all 16 keys into a SIMD register.
	keys := archsimd.LoadInt8x16((*[16]int8)(unsafe.Pointer(&n.Keys)))
	// Broadcast search byte to all 16 lanes.
	cmp := archsimd.BroadcastInt8x16(int8(c))
	// Compare all lanes simultaneously (VPCMPEQB).
	mask := keys.Equal(cmp)
	// Extract one bit per lane into a uint16 (VPMOVMSKB).
	bitmask := mask.ToBits()
	// Mask off slots beyond count.
	bitmask &= (1 << n.Count) - 1

	if bitmask == 0 {
		return nil
	}
	idx := bits.TrailingZeros16(bitmask)
	return &n.Children[idx]
}

func (n *Node16[V]) AddChild(c byte, child *Node[V]) {
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

func (n *Node16[V]) RemoveChild(c byte) {
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
