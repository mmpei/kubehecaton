package common

import (
	"hash/fnv"
	"sort"
)

const (
	a = 1103515245    // multiplier
	c = 12345         // increment
	m = (1 << 31) - 1 // modulus (2**32-1)
)

// SortByWeight returns the given set of nodes sorted in decreasing order of
// their weight for the given key.
func SortByWeight(nodes []string, key []byte) []string {
	d := hash(key)

	entries := make(entryList, len(nodes))

	for i, node := range nodes {
		entries[i] = entry{node: node, weight: weight(node, d)}
	}

	sort.Sort(entries)

	sorted := make([]string, len(entries))
	for i, e := range entries {
		sorted[i] = e.node
	}
	return sorted
}

func weight(s string, d int32) int {
	hs := hash([]byte(s))
	v := (a * ((a*hs + c) ^ d + c))
	if v < 0 {
		v += m
	}
	return int(v)
}

type entry struct {
	node   string
	weight int
}

type entryList []entry

func (l entryList) Len() int {
	return len(l)
}

func (l entryList) Less(a, b int) bool {
	return l[a].weight > l[b].weight
}

func (l entryList) Swap(a, b int) {
	l[a], l[b] = l[b], l[a]
}

func hash(b []byte) int32 {
	h := fnv.New32a()
	h.Write(b)
	return int32(h.Sum32())
}
