package merkledag

option (gogoproto.gostring_all) = true
option (gogoproto.equal_all) = true
option (gogoproto.verbose_equal_all) = true
option (gogoproto.goproto_stringer_all) = false
option (gogoproto.stringer_all) = true
option (gogoproto.populate_all) = true
option (gogoproto.testgen_all) = true
option (gogoproto.benchgen_all) = true
option (gogoproto.marshaler_all) = true
option (gogoproto.sizer_all) = true
option (gogoproto.unmarshaler_all) = true

// An IPFS MerkleDAG Link
message PBLink {

// multihash of the target object
optional bytes Hash = 1;

// utf string name. should be unique per object
optional string Name = 2;

// cumulative size of target object
optional uint64 TSize = 3;
}
