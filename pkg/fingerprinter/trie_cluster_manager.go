// Copyright 2024-2025 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fingerprinter

import (
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
)

// cluster holds one bucket of previously-seen TokenSeq.Items
// and maintains Jaccard statistics over time.
type cluster struct {
	fingerprint int64
	tokenSet    map[string]struct{}
	matchCount  int
	total       int
	matchRate   float64
	lastUpdated time.Time
}

// record updates this cluster’s stats with a new incoming set.
// If matched==true, it also intersects tokenSet to common tokens.
func (c *cluster) record(incoming map[string]struct{}, matched bool) {
	if matched {
		common := make(map[string]struct{}, len(c.tokenSet))
		for tok := range c.tokenSet {
			if _, ok := incoming[tok]; ok {
				common[tok] = struct{}{}
			}
		}
		c.tokenSet = common
		c.matchCount++
	}
	c.total++
	c.matchRate = float64(c.matchCount) / float64(c.total)
	c.lastUpdated = time.Now()
}

// LeafClusterer maintains ordered clusters for one token sequence.
type LeafClusterer struct {
	threshold float64
	mu        sync.Mutex
	clusters  []*cluster // sorted descending by matchRate
}

// NewLeafClusterer creates a clusterer with the given Jaccard threshold.
func NewLeafClusterer(threshold float64) *LeafClusterer {
	return &LeafClusterer{
		threshold: threshold,
		clusters:  make([]*cluster, 0),
	}
}

// Add processes the TokenSeq and returns the canonical fingerprint.
// It matches against existing clusters by Jaccard, updates stats,
// and only computes a new hash if no cluster matches.
func (lc *LeafClusterer) Add(ts *TokenSeq) int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Build incoming set
	incoming := make(map[string]struct{}, len(ts.Items))
	for _, tok := range ts.Items {
		incoming[tok] = struct{}{}
	}

	// Try to match existing clusters
	for idx, cl := range lc.clusters {
		inter := 0
		for tok := range cl.tokenSet {
			if _, ok := incoming[tok]; ok {
				inter++
			}
		}
		union := len(cl.tokenSet) + len(incoming) - inter
		var score float64
		if union > 0 {
			score = float64(inter) / float64(union)
		}
		matched := score >= lc.threshold

		cl.record(incoming, matched)
		if matched {
			// bubble up
			for i := idx; i > 0 && lc.clusters[i].matchRate > lc.clusters[i-1].matchRate; i-- {
				lc.clusters[i], lc.clusters[i-1] = lc.clusters[i-1], lc.clusters[i]
			}
			return cl.fingerprint
		}
	}

	// No match: compute new fingerprint
	h := xxhash.New()
	for i, item := range ts.Items {
		if i > 0 {
			_, _ = h.WriteString(":")
		}
		_, _ = h.WriteString(item)
	}
	for _, key := range ts.JSONKeys {
		_, _ = h.WriteString(":")
		_, _ = h.WriteString(key)
	}
	newFP := int64(h.Sum64())

	// prepend new cluster
	newCl := &cluster{
		fingerprint: newFP,
		tokenSet:    incoming,
		matchCount:  1,
		total:       1,
		matchRate:   1.0,
		lastUpdated: time.Now(),
	}
	lc.clusters = append([]*cluster{newCl}, lc.clusters...)
	return newFP
}

// -----------------------------------------------------------------------------
//  Sequence-trie mapping token sequences -> LeafClusterer
// -----------------------------------------------------------------------------

type seqNode struct {
	children map[string]*seqNode
	cluster  *LeafClusterer
}

func newSeqNode() *seqNode {
	return &seqNode{children: make(map[string]*seqNode)}
}

// TrieClusterManager maps full token prefixes to LeafClusterers.
type TrieClusterManager struct {
	root      *seqNode
	threshold float64
}

// NewTrieClusterManager initializes a manager with the given threshold.
func NewTrieClusterManager(threshold float64) *TrieClusterManager {
	return &TrieClusterManager{
		root:      newSeqNode(),
		threshold: threshold,
	}
}

// getOrCreateLeaf ensures a cluster exists at the current node.
func (m *TrieClusterManager) getOrCreateLeaf(node *seqNode) *LeafClusterer {
	if node.cluster == nil {
		node.cluster = NewLeafClusterer(m.threshold)
	}
	return node.cluster
}

// collectLeafClustersUnder returns every LeafClusterer in the subtree.
func collectLeafClustersUnder(node *seqNode) []*LeafClusterer {
	var out []*LeafClusterer
	var dfs func(n *seqNode)
	dfs = func(n *seqNode) {
		if n.cluster != nil {
			out = append(out, n.cluster)
		}
		for _, c := range n.children {
			dfs(c)
		}
	}
	dfs(node)
	return out
}

// Cluster walks as far as it can down the trie of ts.Items, then
// either exact-adds at leaf, or on divergence Jaccard-scans that subtree.
func (m *TrieClusterManager) Cluster(ts *TokenSeq) int64 {
	cur := m.root
	// 1) walk existing prefix
	var i int
	for i = 0; i < len(ts.Items); i++ {
		tok := ts.Items[i]
		nxt, ok := cur.children[tok]
		if !ok {
			break
		}
		cur = nxt
	}

	// 2) exact match
	if i == len(ts.Items) {
		leaf := m.getOrCreateLeaf(cur)
		return leaf.Add(ts)
	}

	// 3) divergence: collect clusters under 'cur'
	incoming := make(map[string]struct{}, len(ts.Items))
	for _, tok := range ts.Items {
		incoming[tok] = struct{}{}
	}
	bestScore := -1.0
	var bestCl *cluster
	var bestLeaf *LeafClusterer

	candidates := collectLeafClustersUnder(cur)
	for _, leaf := range candidates {
		leaf.mu.Lock()
		for _, cl := range leaf.clusters {
			// run Jaccard Similarity on leaf.clusters
			inter := 0
			for tok := range cl.tokenSet {
				if _, ok := incoming[tok]; ok {
					inter++
				}
			}
			union := len(cl.tokenSet) + len(incoming) - inter
			var score float64
			if union > 0 {
				score = float64(inter) / float64(union)
			}
			if score >= leaf.threshold && score > bestScore {
				bestScore = score
				bestCl = cl
				bestLeaf = leaf
			}
		}
		leaf.mu.Unlock()
	}

	// 4) if a match was found, record & return
	if bestCl != nil && bestLeaf != nil {
		bestLeaf.mu.Lock()
		bestCl.record(incoming, true)
		bestLeaf.mu.Unlock()
		return bestCl.fingerprint
	}

	// 5) otherwise insert the remainder as new path
	for ; i < len(ts.Items); i++ {
		tok := ts.Items[i]
		nxt := newSeqNode()
		cur.children[tok] = nxt
		cur = nxt
	}
	leaf := m.getOrCreateLeaf(cur)
	return leaf.Add(ts)
}
