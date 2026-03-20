# go-simd-art

Adaptive Radix Tree (ART) for Go with SIMD-accelerated Node16 lookups.

Based on [The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases](https://db.in.tum.de/~leis/papers/ART.pdf) (Leis, Kemper, Neumann, 2013).

## Why another Go ART?

Existing Go implementations ([plar/go-adaptive-radix-tree](https://github.com/plar/go-adaptive-radix-tree), [arriqaaq/art](https://github.com/arriqaaq/art), etc.) use linear scan or binary search for Node16 key lookup. The original paper gets a significant speedup from SSE SIMD instructions on Node16, comparing all 16 keys in a single instruction.

Go 1.26 introduced `simd/archsimd` ([proposal](https://github.com/golang/go/issues/73787)), making native SIMD intrinsics available without hand-written assembly. This library uses those intrinsics to implement the Node16 optimization from the paper.

## Status

**Experimental.** This is a from-scratch ART implementation exploring Go's new SIMD capabilities.

## Features

- Node4, Node16 (SIMD), Node48, Node256 adaptive node types
- Insert, Search, Delete
- Prefix scan and longest prefix match
- In-order iteration
- Generic value type (`Tree[V any]`) with zero-allocation lookups
- Zero external dependencies only the Go standard library

## Requirements

- **Go 1.26+** with `GOEXPERIMENT=simd`
- AMD64 (x86-64) only

## Install

```bash
go get github.com/hexfusion/go-simd-art
```

## Build

```bash
GOEXPERIMENT=simd go build ./...
GOEXPERIMENT=simd go test ./...
```

## Usage

```go
import "github.com/hexfusion/go-simd-art/art"

tree := art.New[string]()
tree.Insert([]byte("hello"), "world")

value, found := tree.Search([]byte("hello"))
// value is string, not any no type assertion needed
```

## Benchmarks

AMD Ryzen AI 9 HX 370, Go 1.26.1. Dense benchmarks use short hierarchical keys
(`/a/b/c`) with 16-way branching at each level, maximizing Node16 usage where
SIMD has the most impact.

### SIMD vs Scalar (Node16 FindChild)

| Benchmark | Scalar | SIMD | Improvement |
|-----------|--------|------|-------------|
| SearchDense | 23.99 ns/op | 20.29 ns/op | **-15.4%** |
| SearchDenseMiss | 16.31 ns/op | 15.57 ns/op | -4.5% |
| PrefixScanDense | 50.50 ns/op | 46.80 ns/op | -7.3% |
| LongestPrefixDense | 40.12 ns/op | 37.51 ns/op | -6.5% |
| Search (random keys) | 351.2 ns/op | 302.1 ns/op | -14.0% |
| Insert | 1479 ns/op | 1266 ns/op | -14.4% |
| Delete | 668.6 ns/op | 636.6 ns/op | -4.8% |

All operations are zero-allocation on the lookup path.

Run benchmarks yourself:

```bash
GOEXPERIMENT=simd go test -bench=. -benchmem ./art/...
```

### TODO

Comparative benchmarks against:
- [plar/go-adaptive-radix-tree](https://github.com/plar/go-adaptive-radix-tree)
- [google/btree](https://github.com/google/btree)
- Go built-in `map[string]any`

## References

- [The Adaptive Radix Tree (2013)](https://db.in.tum.de/~leis/papers/ART.pdf): the original paper
- [The ART of Practical Synchronization (2016)](https://db.in.tum.de/~leis/papers/artsync.pdf): ART-OLC concurrency
- [Go SIMD proposal #73787](https://github.com/golang/go/issues/73787): `simd/archsimd` design and discussion
- [simd/archsimd package docs](https://pkg.go.dev/simd/archsimd): `Int8x16`, `Mask8x16.ToBits()` API
- [archsimd source](https://tip.golang.org/src/simd/archsimd/): reference implementation and type definitions
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26): GOEXPERIMENT=simd availability

## License

Apache 2.0
