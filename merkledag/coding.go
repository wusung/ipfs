package merkledag

import (
	"fmt"
	mh "github.com/multiformats/go-multihash"
)

// for now, we use a PBNode intermediate thing.
// because native go objects are nice.

// Unmarshal decodes raw data into a *Node instance.
// The conversion uses an intermediate PBNode.
func (n *Node) Unmarshal(encoded []byte) error {
	var pbn PBNode
	if err := pbn.Unmarshal(encoded); err != nil {
		return fmt.Errorf("Unmarshal failed. %v", err)
	}

	pbnl := pbn.GetLinks()
	n.Links = make([]*Link, len(pbnl))
	for i, l := range pbnl {
		n.Links[i] = &Link{Name: l.GetName(), Size: l.GetTsize()}
		h, err := mh.Cast(l.GetHash())
		if err != nil {
			return fmt.Errorf("Link hash is not valid multihash. %v", err)
		}
		n.Links[i].Hash = h
	}

	n.Data = pbn.GetData()
	return nil
}

// MarshalTo encodes a *Node instance into a given byte slice.
// The conversion uses an intermediate PBNode.
func (n *Node) MarshalTo(encoded []byte) error {
	pbn := n.getPBNode()
	if _, err := pbn.MarshalTo(encoded); err != nil {
		return fmt.Errorf("Marshal failed. %v", err)
	}
	return nil
}

// Marshal encodes a *Node instance into a new byte slice.
// The conversion uses an intermediate PBNode.
func (n *Node) Marshal() ([]byte, error) {
	pbn := n.getPBNode()
	data, err := pbn.Marshal()
	if err != nil {
		return data, fmt.Errorf("Marshal failed. %v", err)
	}
	return data, nil
}

func (n *Node) getPBNode() *PBNode {
	pbn := &PBNode{}
	pbn.Links = make([]*PBLink, len(n.Links))
	for i, l := range n.Links {
		pbn.Links[i] = &PBLink{}
		pbn.Links[i].Name = &l.Name
		pbn.Links[i].Tsize = &l.Size
		pbn.Links[i].Hash = []byte(l.Hash)
	}

	pbn.Data = n.Data
	return pbn
}

// Encoded returns the encoded raw data version of a Node instance.
// It may use a cached encoded version, unless the force flag is given.
func (n *Node) Encoded(force bool) ([]byte, error) {
	if n.encoded == nil || force {
		var err error
		n.encoded, err = n.Marshal()
		if err != nil {
			return []byte{}, err
		}
	}

	return n.encoded, nil
}

// Decoded decodes raw data and returns a new Node instance.
func Decoded(encoded []byte) (*Node, error) {
	n := &Node{}
	err := n.Unmarshal(encoded)
	return n, err
}
