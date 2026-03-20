package art

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"testing"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		search string
		want   int
		found  bool
	}{
		{"hit", []string{"apple", "app", "application", "banana"}, "app", 1, true},
		{"miss", []string{"apple", "banana"}, "missing", 0, false},
		{"single key", []string{"only"}, "only", 0, true},
		{"prefix of existing", []string{"abcdef"}, "abc", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[int]()
			for i, k := range tc.keys {
				tree.Insert([]byte(k), i)
			}
			val, found := tree.Search([]byte(tc.search))
			if found != tc.found {
				t.Fatalf("found=%v, want %v", found, tc.found)
			}
			if found && val != tc.want {
				t.Fatalf("val=%d, want %d", val, tc.want)
			}
		})
	}
}

func TestInsertReplace(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		v1, v2   string
		wantOld  string
		replaced bool
	}{
		{"replace existing", "key", "v1", "v2", "v1", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[string]()
			tree.Insert([]byte(tc.key), tc.v1)
			old, replaced := tree.Insert([]byte(tc.key), tc.v2)
			if replaced != tc.replaced || old != tc.wantOld {
				t.Fatalf("old=%q replaced=%v, want %q %v", old, replaced, tc.wantOld, tc.replaced)
			}
			if tree.Size() != 1 {
				t.Fatalf("size=%d, want 1", tree.Size())
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		delete    string
		deleted   bool
		remaining []string
	}{
		{"existing key", []string{"foo", "foobar", "bar"}, "foobar", true, []string{"foo", "bar"}},
		{"non-existent", []string{"foo", "bar"}, "nope", false, []string{"foo", "bar"}},
		{"single key", []string{"only"}, "only", true, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[string]()
			for _, k := range tc.keys {
				tree.Insert([]byte(k), k)
			}
			_, deleted := tree.Delete([]byte(tc.delete))
			if deleted != tc.deleted {
				t.Fatalf("deleted=%v, want %v", deleted, tc.deleted)
			}
			for _, k := range tc.remaining {
				if _, found := tree.Search([]byte(k)); !found {
					t.Fatalf("key %q should still exist", k)
				}
			}
		})
	}
}

func TestDeleteAll(t *testing.T) {
	keys := []string{"a", "ab", "abc", "abcd", "b", "bc", "bcd"}
	tree := New[int]()
	for i, k := range keys {
		tree.Insert([]byte(k), i)
	}
	for _, k := range keys {
		if _, deleted := tree.Delete([]byte(k)); !deleted {
			t.Fatalf("failed to delete %q", k)
		}
	}
	if tree.Size() != 0 {
		t.Fatalf("size=%d, want 0", tree.Size())
	}
	if _, found := tree.Search([]byte("a")); found {
		t.Fatal("tree should be empty")
	}
}

func TestForEachSorted(t *testing.T) {
	tree := New[string]()
	keys := []string{"charlie", "alpha", "bravo", "delta", "echo"}
	for _, k := range keys {
		tree.Insert([]byte(k), k)
	}

	var result []string
	tree.ForEach(func(key []byte, _ string) bool {
		result = append(result, string(key))
		return true
	})

	sort.Strings(keys)
	for i := range keys {
		if result[i] != keys[i] {
			t.Fatalf("position %d: got %q, want %q", i, result[i], keys[i])
		}
	}
}

func TestForEachEarlyStop(t *testing.T) {
	tree := New[int]()
	for i := 0; i < 100; i++ {
		tree.Insert([]byte(fmt.Sprintf("key/%03d", i)), i)
	}
	count := 0
	tree.ForEach(func(_ []byte, _ int) bool {
		count++
		return count < 5
	})
	if count != 5 {
		t.Fatalf("count=%d, want 5", count)
	}
}

func TestForEachPrefix(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		prefix   string
		expected []string
	}{
		{"match", []string{"test", "testing", "tester", "team", "toast"}, "test", []string{"test", "tester", "testing"}},
		{"no match", []string{"hello"}, "xyz", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[string]()
			for _, w := range tc.keys {
				tree.Insert([]byte(w), w)
			}
			var result []string
			tree.ForEachPrefix([]byte(tc.prefix), func(key []byte, _ string) bool {
				result = append(result, string(key))
				return true
			})
			sort.Strings(result)
			if len(result) != len(tc.expected) {
				t.Fatalf("got %v, want %v", result, tc.expected)
			}
			for i := range tc.expected {
				if result[i] != tc.expected[i] {
					t.Fatalf("position %d: got %q, want %q", i, result[i], tc.expected[i])
				}
			}
		})
	}
}

func TestLongestPrefix(t *testing.T) {
	tree := New[string]()
	for _, p := range []string{"/", "/api", "/api/v1", "/api/v1/users"} {
		tree.Insert([]byte(p), p)
	}

	tests := []struct {
		key  string
		want string
		found bool
	}{
		{"/api/v1/users/123", "/api/v1/users", true},
		{"/api/v1/posts", "/api/v1", true},
		{"/api/v2", "/api", true},
		{"/other", "/", true},
		{"/api/v1", "/api/v1", true},
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			_, val, found := tree.LongestPrefix([]byte(tc.key))
			if found != tc.found {
				t.Fatalf("found=%v, want %v", found, tc.found)
			}
			if found && val != tc.want {
				t.Fatalf("val=%q, want %q", val, tc.want)
			}
		})
	}
}

func TestLongestPrefixNoMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert([]byte("/api/v1"), "v1")
	if _, _, found := tree.LongestPrefix([]byte("/other")); found {
		t.Fatal("expected no match")
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		wantMin  string
		wantMax  string
	}{
		{"basic", []string{"mango", "apple", "zebra", "banana"}, "apple", "zebra"},
		{"single", []string{"only"}, "only", "only"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree := New[string]()
			for _, k := range tc.keys {
				tree.Insert([]byte(k), k)
			}
			key, _, found := tree.Minimum()
			if !found || string(key) != tc.wantMin {
				t.Fatalf("min=%q, want %q", key, tc.wantMin)
			}
			key, _, found = tree.Maximum()
			if !found || string(key) != tc.wantMax {
				t.Fatalf("max=%q, want %q", key, tc.wantMax)
			}
		})
	}
}

func TestEmptyTree(t *testing.T) {
	tree := New[string]()
	if _, found := tree.Search([]byte("x")); found {
		t.Fatal("search should miss")
	}
	if _, deleted := tree.Delete([]byte("x")); deleted {
		t.Fatal("delete should miss")
	}
	if _, _, found := tree.Minimum(); found {
		t.Fatal("min should miss")
	}
	if _, _, found := tree.Maximum(); found {
		t.Fatal("max should miss")
	}
	if tree.Size() != 0 {
		t.Fatalf("size=%d, want 0", tree.Size())
	}
}

func TestEmptyKey(t *testing.T) {
	tree := New[int]()
	tree.Insert([]byte{}, 42)
	tree.Insert([]byte("a"), 1)

	val, found := tree.Search([]byte{})
	if !found || val != 42 {
		t.Fatalf("val=%v found=%v, want 42 true", val, found)
	}
	val, found = tree.Search([]byte("a"))
	if !found || val != 1 {
		t.Fatalf("val=%v found=%v, want 1 true", val, found)
	}
}

func TestBinaryKeys(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{"low bytes", []byte{0x00, 0x01, 0x02}},
		{"high bytes", []byte{0xFF, 0xFE, 0xFD}},
		{"mixed", []byte{0x00, 0x01, 0xFF}},
		{"single 0x80", []byte{0x80}},
		{"single 0x00", []byte{0x00}},
	}
	tree := New[int]()
	for i, tc := range tests {
		tree.Insert(tc.key, i)
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, found := tree.Search(tc.key)
			if !found || val != i {
				t.Fatalf("val=%d found=%v, want %d true", val, found, i)
			}
		})
	}
}

func TestNullBytesInKeys(t *testing.T) {
	tests := []struct {
		key  string
		val  string
	}{
		{"a\x00b", "a-null-b"},
		{"a\x00c", "a-null-c"},
		{"a\x00", "a-null"},
	}
	tree := New[string]()
	for _, tc := range tests {
		tree.Insert([]byte(tc.key), tc.val)
	}
	for _, tc := range tests {
		t.Run(tc.val, func(t *testing.T) {
			val, found := tree.Search([]byte(tc.key))
			if !found || val != tc.val {
				t.Fatalf("val=%q found=%v, want %q true", val, found, tc.val)
			}
		})
	}
}

func TestPrefixKeyOrdering(t *testing.T) {
	tests := []struct {
		key string
		val int
	}{
		{"abcdef", 1},
		{"abc", 2},
		{"abcdefgh", 3},
	}
	tree := New[int]()
	for _, tc := range tests {
		tree.Insert([]byte(tc.key), tc.val)
	}
	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			val, found := tree.Search([]byte(tc.key))
			if !found || val != tc.val {
				t.Fatalf("val=%d found=%v, want %d true", val, found, tc.val)
			}
		})
	}
}

func TestNodeGrowth(t *testing.T) {
	tree := New[int]()
	for i := 0; i < 256; i++ {
		tree.Insert([]byte(fmt.Sprintf("key_%02x", i)), i)
	}
	if tree.Size() != 256 {
		t.Fatalf("size=%d, want 256", tree.Size())
	}
	for i := 0; i < 256; i++ {
		key := fmt.Sprintf("key_%02x", i)
		val, found := tree.Search([]byte(key))
		if !found || val != i {
			t.Fatalf("key %q: val=%d found=%v, want %d true", key, val, found, i)
		}
	}
}

func TestNodeShrinking(t *testing.T) {
	tree := New[int]()
	var keys [][]byte
	for i := 0; i < 48; i++ {
		key := []byte(fmt.Sprintf("key/%c", rune('A'+i)))
		keys = append(keys, key)
		tree.Insert(key, i)
	}

	// Delete 44 to trigger Node48 -> Node16 -> Node4.
	for i := 0; i < 44; i++ {
		if _, deleted := tree.Delete(keys[i]); !deleted {
			t.Fatalf("failed to delete %q", keys[i])
		}
	}
	if tree.Size() != 4 {
		t.Fatalf("size=%d, want 4", tree.Size())
	}
	for i := 44; i < 48; i++ {
		val, found := tree.Search(keys[i])
		if !found || val != i {
			t.Fatalf("key %q: val=%d found=%v, want %d true", keys[i], val, found, i)
		}
	}
}

func TestLargeScaleConsistency(t *testing.T) {
	tree := New[int]()
	n := 10000
	keys := make([][]byte, n)
	for i := range keys {
		keys[i] = []byte(fmt.Sprintf("item/%05d/data", i))
	}
	for i, k := range keys {
		tree.Insert(k, i)
	}

	// Delete odd-indexed keys, verify even present and odd absent.
	for i := 1; i < n; i += 2 {
		if _, deleted := tree.Delete(keys[i]); !deleted {
			t.Fatalf("failed to delete %q", keys[i])
		}
	}
	if tree.Size() != n/2 {
		t.Fatalf("size=%d, want %d", tree.Size(), n/2)
	}
	for i, k := range keys {
		_, found := tree.Search(k)
		if i%2 == 0 && !found {
			t.Fatalf("even key %q should exist", k)
		}
		if i%2 == 1 && found {
			t.Fatalf("odd key %q should be deleted", k)
		}
	}
}

// Benchmarks

func BenchmarkInsert(b *testing.B) {
	keys := generateKeys(b.N)
	b.ResetTimer()
	tree := New[int]()
	for i := 0; i < b.N; i++ {
		tree.Insert(keys[i], i)
	}
}

func BenchmarkSearch(b *testing.B) {
	tree := New[int]()
	keys := generateKeys(100000)
	for i, k := range keys {
		tree.Insert(k, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search(keys[i%len(keys)])
	}
}

func BenchmarkSearchMiss(b *testing.B) {
	tree := New[int]()
	keys := generateKeys(100000)
	for i, k := range keys {
		tree.Insert(k, i)
	}
	missKeys := make([][]byte, 100000)
	for i := range missKeys {
		missKeys[i] = []byte(fmt.Sprintf("miss_%016x", rand.Int64()))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search(missKeys[i%len(missKeys)])
	}
}

func BenchmarkDelete(b *testing.B) {
	keys := generateKeys(b.N)
	tree := New[int]()
	for i, k := range keys {
		tree.Insert(k, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Delete(keys[i])
	}
}

// Dense benchmarks: short keys with 16-way branching to maximize Node16 SIMD.
func BenchmarkSearchDense(b *testing.B) {
	tree := New[int]()
	var keys [][]byte
	for a := byte('a'); a <= 'p'; a++ {
		for c := byte('a'); c <= 'p'; c++ {
			for d := byte('a'); d <= 'p'; d++ {
				key := []byte{'/', a, '/', c, '/', d}
				keys = append(keys, key)
				tree.Insert(key, int(a)*256+int(c)*16+int(d))
			}
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search(keys[i%len(keys)])
	}
}

func BenchmarkSearchDenseMiss(b *testing.B) {
	tree := New[int]()
	for a := byte('a'); a <= 'p'; a++ {
		for c := byte('a'); c <= 'p'; c++ {
			for d := byte('a'); d <= 'p'; d++ {
				tree.Insert([]byte{'/', a, '/', c, '/', d}, 0)
			}
		}
	}
	var missKeys [][]byte
	for a := byte('a'); a <= 'p'; a++ {
		for c := byte('a'); c <= 'p'; c++ {
			missKeys = append(missKeys, []byte{'/', a, '/', c, '/', 'z'})
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search(missKeys[i%len(missKeys)])
	}
}

func BenchmarkPrefixScanDense(b *testing.B) {
	tree := New[int]()
	for a := byte('a'); a <= 'p'; a++ {
		for c := byte('a'); c <= 'p'; c++ {
			for d := byte('a'); d <= 'p'; d++ {
				tree.Insert([]byte{'/', a, '/', c, '/', d}, 0)
			}
		}
	}
	prefix := []byte("/h/h")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.ForEachPrefix(prefix, func(_ []byte, _ int) bool { return true })
	}
}

func BenchmarkLongestPrefixDense(b *testing.B) {
	tree := New[string]()
	for _, r := range []string{"/", "/a", "/a/b", "/a/b/c", "/a/b/c/d", "/a/b/c/d/e"} {
		tree.Insert([]byte(r), r)
	}
	lookup := []byte("/a/b/c/d/e/f/g")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.LongestPrefix(lookup)
	}
}

func generateKeys(n int) [][]byte {
	keys := make([][]byte, n)
	for i := range keys {
		keys[i] = []byte(fmt.Sprintf("key_%016x", rand.Int64()))
	}
	return keys
}
