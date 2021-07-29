package times

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"

	"github.com/carterpeel/go-corelib/jsons"
	"github.com/carterpeel/go-corelib/strs"
)

func TestAddToTrie(t *testing.T) {
	trie := strs.NewRuneTrie()
	addToTrie(trie, trieEntry{Pattern: "abc", Layout: "123"})
	assert.PanicsWithValue(t,
		"pattern 'abc' caused a collision",
		func() {
			addToTrie(trie, trieEntry{Pattern: "abc", Layout: "dup"})
		})
}

func TestInitDateTimeTrie(t *testing.T) {
	trie := initDateTimeTrie()
	t.Logf("total trie nodes created: %d", trie.NodeCount())
	cupaloy.SnapshotT(t, jsons.BPM(trie))
}

func TestInitTimezones(t *testing.T) {
	m := initTimezones()
	t.Logf("total timezones: %d", len(m))
	cupaloy.SnapshotT(t, jsons.BPM(m))
}

// BenchmarkInitDateTimeTrie-8   	     100	  18147722 ns/op	 6505165 B/op	  129455 allocs/op
func BenchmarkInitDateTimeTrie(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = initDateTimeTrie()
	}
}

// BenchmarkInitTimezones-8      	   10000	    109394 ns/op	   45532 B/op	      24 allocs/op
func BenchmarkInitTimezones(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = initTimezones()
	}
}
