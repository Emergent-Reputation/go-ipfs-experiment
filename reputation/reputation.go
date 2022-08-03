package reputation

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

func BuildFriendListNode(friends []string) datamodel.Node {
	size := len(friends)
	friendList, err := qp.BuildList(basicnode.Prototype.List, int64(size), func(assembler datamodel.ListAssembler) {
		for _, v := range friends {
			qp.ListEntry(assembler, qp.String(v))

		}
	})

	// TODO:(@ckartik) handle gracefully
	if err != nil {
		panic(err)
	}

	return friendList
}
