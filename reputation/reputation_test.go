package reputation

import (
	"testing"

	ipld "github.com/ipld/go-ipld-prime"
	cbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
)

func Test_tmp(t *testing.T) {
	ipld.Marshal(cbor.Encode, struct{ string }{string: "Kartik"})

}
