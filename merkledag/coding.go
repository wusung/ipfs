package merkledag

import (
	"fmt"
	mh "github.com/jbenet/go-multihash"
)

// for now, we use a PBNode intermediate thing.
// because native go objects are nice.

func (n *Node) Unmarshal(encoded []byte) error {
	var pbn PBNode
	if err := pbn.Unmarshal(encoded); err != nil {
		return fmt.Errorf("Unmarshal failed. %v", err)
	}

	pbnl := pbn.GetLinks()
	n.Links = make([]*Link, len(pbnl))
	for i, l := range (pbnl) {
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

func (n *Node) MarshalTo(encoded []byte) error {
	pbn := n.getPBNode()
	if _, err := pbn.MarshalTo(encoded); err != nil {
		return fmt.Errorf("Marshal failed. %v", err)
	}
	return nil
}

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
	for i, l := range (n.Links) {
		pbn.Links[i] = &PBLink{}
		pbn.Links[i].Name = &l.Name
		pbn.Links[i].Tsize = &l.Size
		pbn.Links[i].Hash = []byte(l.Hash)
	}

	pbn.Data = n.Data
	return pbn
}

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

func Decoded(encoded []byte) (*Node, error) {
	n := &Node{}
	err := n.Unmarshal(encoded)
	return n, err
}
