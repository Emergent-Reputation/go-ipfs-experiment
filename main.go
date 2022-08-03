package main

import (
	"fmt"

	ipld "github.com/ipld/go-ipld-prime"
  cbor "github.com/ipld/go-ipld-prime/codec/dagcbor"
)

func main() {

	fmt.Println(cid)
  
	ipld.Marshal(cbor.Encoder, struct { var string } {string:"Kartik"})
	fmt.Println("Hello")
}
