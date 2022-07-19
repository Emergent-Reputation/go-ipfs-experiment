package main

import (
	"fmt"
	"strings"

	ipfs "github.com/ipfs/go-ipfs-api"
)

func main() {
	shell := ipfs.NewShell("localhost:5001")
	cid, err := shell.Add(strings.NewReader("hello"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cid)
	fmt.Println("Hello")
}
