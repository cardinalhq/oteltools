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
)

// ----------------------------------------------------------------------------
// internal cluster & leaf clusterer
// ----------------------------------------------------------------------------

type cluster struct {
	Fingerprint int64               `json:"fp"`
	TokenSet    map[string]struct{} `json:"tokens"`
	MatchCount  int                 `json:"mc"`
	Total       int                 `json:"tot"`
	LastUpdated time.Time           `json:"upd"`
}

func (c *cluster) matchRate() float64 {
	if c.Total == 0 {
		return 0
	}
	return float64(c.MatchCount) / float64(c.Total)
}

func (c *cluster) record(incoming map[string]struct{}, matched bool) {
	if matched {
		// intersect token set
		common := make(map[string]struct{}, len(c.TokenSet))
		for tok := range c.TokenSet {
			if _, ok := incoming[tok]; ok {
				common[tok] = struct{}{}
			}
		}
		c.TokenSet = common
		c.MatchCount++
	}
	c.Total++
	c.LastUpdated = time.Now()
}

// LeafClusterer keeps ordered clusters for one token-prefix node.
type LeafClusterer struct {
	threshold float64
	mu        sync.Mutex
	clusters  []*cluster // descending by match rate
}

func NewLeafClusterer(threshold float64) *LeafClusterer {
	return &LeafClusterer{
		threshold: threshold,
		clusters:  make([]*cluster, 0),
	}
}

// Add returns either an existing cluster fingerprint (if jaccard >= threshold)
// or computes & returns a new one.
func (lc *LeafClusterer) Add(ts *TokenSeq) int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// build incoming set
	incoming := make(map[string]struct{}, len(ts.Items))
	for _, tok := range ts.Items {
		incoming[tok] = struct{}{}
	}

	// try to match
	for idx, cl := range lc.clusters {
		// Jaccard
		inter := 0
		for tok := range cl.TokenSet {
			if _, ok := incoming[tok]; ok {
				inter++
			}
		}
		union := len(cl.TokenSet) + len(incoming) - inter
		score := 0.0
		if union > 0 {
			score = float64(inter) / float64(union)
		}

		cl.record(incoming, score >= lc.threshold)
		if score >= lc.threshold {
			// bubble up
			for i := idx; i > 0 && lc.clusters[i].matchRate() > lc.clusters[i-1].matchRate(); i-- {
				lc.clusters[i], lc.clusters[i-1] = lc.clusters[i-1], lc.clusters[i]
			}
			return cl.Fingerprint
		}
	}

	// no match → new fingerprint
	newFP := fpr.FingerprintItemsAndJSONKeys(ts)

	cl := &cluster{
		Fingerprint: newFP,
		TokenSet:    incoming,
		MatchCount:  1,
		Total:       1,
		LastUpdated: time.Now(),
	}
	lc.clusters = append([]*cluster{cl}, lc.clusters...)
	return newFP
}

// ----------------------------------------------------------------------------
// trie + manager + serialization
// ----------------------------------------------------------------------------

type seqNode struct {
	children map[string]*seqNode
	leaf     *LeafClusterer
	sync.RWMutex
}

func newSeqNode() *seqNode {
	return &seqNode{children: make(map[string]*seqNode)}
}

// TrieClusterManager maps full token-prefixes to per-node LeafClusterers.
type TrieClusterManager struct {
	root      *seqNode
	threshold float64
}

func NewTrieClusterManager(threshold float64) *TrieClusterManager {
	return &TrieClusterManager{
		root:      newSeqNode(),
		threshold: threshold,
	}
}

func (m *TrieClusterManager) getOrCreateLeaf(n *seqNode) *LeafClusterer {
	n.Lock()
	defer n.Unlock()

	if n.leaf == nil {
		n.leaf = NewLeafClusterer(m.threshold)
	}
	return n.leaf
}

func collectLeafers(n *seqNode) []*LeafClusterer {
	var out []*LeafClusterer
	var dfs func(x *seqNode)
	dfs = func(x *seqNode) {
		x.RLock()
		if x.leaf != nil {
			out = append(out, x.leaf)
		}
		// snapshot children pointers under the same lock
		children := make([]*seqNode, 0, len(x.children))
		for _, c := range x.children {
			children = append(children, c)
		}
		x.RUnlock()

		// recurse after unlocking
		for _, c := range children {
			dfs(c)
		}
	}
	dfs(n)
	return out
}

// Cluster walks as far down ts.Items as possible in the trie.
// If it consumes all tokens → exact leaf.Add.
// On divergence, it Jaccard-scans every cluster under the current subtree,
// picks the best above threshold, or else creates a new branch + leaf.
func (m *TrieClusterManager) Cluster(ts *TokenSeq) int64 {
	cur := m.root
	i := 0
	for ; i < len(ts.Items); i++ {
		cur.RLock()
		nxt, ok := cur.children[ts.Items[i]]
		cur.RUnlock()
		if !ok {
			break
		}
		cur = nxt
	}

	// exact
	if i == len(ts.Items) {
		return m.getOrCreateLeaf(cur).Add(ts)
	}

	// divergence → scan under cur
	incoming := make(map[string]struct{}, len(ts.Items))
	for _, tok := range ts.Items {
		incoming[tok] = struct{}{}
	}

	bestScore := -1.0
	var bestCl *cluster
	var bestLeaf *LeafClusterer
	for _, leaf := range collectLeafers(cur) {
		leaf.mu.Lock()
		for _, cl := range leaf.clusters {
			inter := 0
			for tok := range cl.TokenSet {
				if _, ok := incoming[tok]; ok {
					inter++
				}
			}
			union := len(cl.TokenSet) + len(incoming) - inter
			score := 0.0
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

	if bestLeaf != nil && bestCl != nil {
		bestLeaf.mu.Lock()
		bestCl.record(incoming, true)
		bestLeaf.mu.Unlock()
		return bestCl.Fingerprint
	}

	// no match → carve out the remainder
	for ; i < len(ts.Items); i++ {
		n := newSeqNode()
		cur.Lock()
		cur.children[ts.Items[i]] = n
		cur.Unlock()
		cur = n
	}
	return m.getOrCreateLeaf(cur).Add(ts)
}
