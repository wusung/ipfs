package dht

import (
	"encoding/hex"
	"container/list"
	"sort"

	peer "../../peer"
	"golang.org/x/net/html/atom"
	"github.com/libp2p/go-libp2p-peerstore"
	"bytes"

	u "../../util"
)

// RoutingTable defines the routing table.
type RoutingTable struct {

	// ID of the local peer
	local ID

	// kBuckets define all the fingers to other nodes.
	Buckets []*Bucket
	bucketsize int
}

func NewRoutingTable(bucketsize int, local_id ID) *RoutingTable {
	rt := new(RoutingTable)
	rt.Buckets = []*Bucket{new(Bucket)}
	rt.bucketsize = bucketsize
	rt.local = local_id
	return rt
}

// Update adds or moves the given peer to the front of its respective bucket
// If a peer gets removed from a bucket, it is returned
func (rt *RoutingTable) Update(p *peer.Peer) *peer.Peer {
	peer_id := convertPeerID(p.ID)
	cpl := xor(peer_id, rt.local).commonPrefixLen()

	b_id := cpl
	if b_id >= len(rt.Buckets) {
		b_id = len(rt.Buckets) - 1
	}

	bucket := rt.Buckets[b_id]
	e := bucket.Find(p.ID)
	if e == nil {
		// New peer, add to bucket
		bucket.PushFront(e)

		if bucket.Len() > rt.bucketsize {
			if b_id == len(rt.Buckets) - 1 {
				new_bucket := bucket.Split(b_id, rt.local)
				if new_bucket.Len() > rt.bucketsize {
					// TODO: This is a very rare and annoying case
					panic("Case not handled.")
				}
				rt.Buckets = append(rt.Buckets, new_bucket)

				// If all elements were on left side of split...
				if bucket.Len() > rt.bucketsize {
					return bucket.PopBack()
				}
			} else {
				// If the bucket cant split kick out least active node
				return bucket.PopBack()
			}
		}
		return nil
	} else {
		// If the peer is already in the table, move it to the front.
		// This signifies that it it "more active" and the less active nodes
		// Will as a result tend towards the back of the list
		bucket.MoveToFront(e)
		return nil
	}
}

// A helper struct to sort peers by their distance to the local node
type peerDistance struct {
	p *peer.Peer
	distance ID
}

// peerSorterArr implements sort.Interface to sort peers by xor distance
type peerSorterArr []*peerDistance
func (p peerSorterArr) Len() int {return len(p)}
func (p peerSorterArr) Swap(a, b int) {p[a],p[b] = p[b],p[a]}
func (p peerSorterArr) Less(a, b int) bool {
	return p[a].distance.Less(p[b].distance)
}

func copyPeersFromList(target ID, peerArr peerSorterArr, peerList *list.List) peerSorterArr {
	for e := peerList.Front(); e != nil; e = e.Next() {
		p := e.Value.(*peer.Peer)
		p_id := convertPeerID(p.ID)
		pd := peerDistance{
			p: p,
			distance: xor(target, p_id),
		}
		peerArr = append(peerArr, &pd)
	}
	return peerArr
}

// Returns a single peer that is nearest to the given ID
func (rt *RoutingTable) NearestPeer(id ID) *peer.Peer {
	peers := rt.NearestPeers(id, 1)
	if len(peers) > 0 {
		return peers[0]
	} else {
		return nil
	}
}

// Returns a list of the 'count' closest peers to the given ID
func (rt *RoutingTable) NearestPeers(id ID, count int) []*peer.Peer {
	u.POut("Searching table, size = %d", rt.Size())
	cpl := xor(id, rt.local).commonPrefixLen()

	// Get bucket at cpl index or last bucket
	var bucket *Bucket
	if cpl >= len(rt.Buckets) {
		cpl = len(rt.Buckets) - 1
	}
	bucket = rt.Buckets[cpl]

	var peerArr peerSorterArr
	if bucket.Len() == 0 {
		// In the case of an unusual split, one bucket may be empty.
		// if this happens, search both surrounding buckets for nearest peer
		if cpl > 0 {
			plist := (*list.List)(rt.Buckets[cpl - 1])
			peerArr = copyPeersFromList(id, peerArr, plist)
		}

		if cpl < len(rt.Buckets) - 1 {
			plist := (*list.List)(rt.Buckets[cpl + 1])
			peerArr = copyPeersFromList(id, peerArr, plist)
		}
	} else {
		plist := (*list.List)(bucket)
		peerArr = copyPeersFromList(id, peerArr, plist)
	}

	// Sort by distance to local peer
	sort.Sort(peerArr)

	var out []*peer.Peer
	for i := 0; i < count && i < peerArr.Len(); i++ {
		out = append(out, peerArr[i].p)
		u.POut("peer out: %s - %s", peerArr[i].p.ID.Pretty(),
		hex.EncodeToString(xor(id, convertPeerID(peerArr[i].p.ID))))
	}

	return out
}

//TODO: make this accept an ID, requires method of converting keys to IDs
func (rt *RoutingTable) NearestNode(key u.Key) *peer.Peer {
	panic("Function not implemented.")
}

func (rt *RoutingTable Size() int {
	var tot int
	for _, buck := range rt.Buckets {
		tot += buck.Len()
	}
	return tot
}
