package reputation

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/memstore"

	"os"
	"testing"
)

var (
	standardInput = []string{"test-1", "test-2", "test-3"}
)

func Test_list_building(t *testing.T) {
	friendList := BuildFriendListNode(standardInput)
	dagjson.Encode(friendList, os.Stdout)
}

var store = memstore.Store{}

func Test_serialize_and_store(t *testing.T) {
	lsys := cidlink.DefaultLinkSystem()

	lsys.SetWriteStorage(&store)
	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,           // Usually '1'.
		Codec:    cid.DagCBOR, // TODO(@ckartik): move this to cbor https://github.com/multiformats/multicodec/
		MhType:   0x13,        // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 64,          // sha2-512 hash has a 64-byte sum.
	}}
	list := BuildFriendListNode(standardInput)
	fmt.Println(lsys)

	lnk, err := lsys.Store(
		linking.LinkContext{},
		lp,
		list,
	)
	if err != nil {
		panic(err)
	}
	t.Log(lnk)
}
