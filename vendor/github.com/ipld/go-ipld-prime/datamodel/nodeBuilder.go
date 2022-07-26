package datamodel

// NodeAssembler is the interface that describes all the ways we can set values
// in a node that's under construction.
//
// A NodeAssembler is about filling in data.
// To create a new Node, you should start with a NodeBuilder (which contains a
// superset of the NodeAssembler methods, and can return the finished Node
// from its `Build` method).
// While continuing to build a recursive structure from there,
// you'll see NodeAssembler for all the child values.
//
// For filling scalar data, there's a `Assign{Kind}` method for each kind;
// after calling one of these methods, the data is filled in, and the assembler is done.
// For recursives, there are `BeginMap` and `BeginList` methods,
// which return an object that needs further manipulation to fill in the contents.
//
// There is also one special method: `AssignNode`.
// `AssignNode` takes another `Node` as a parameter,
// and should should internally call one of the other `Assign*` or `Begin*` (and subsequent) functions
// as appropriate for the kind of the `Node` it is given.
// This is roughly equivalent to using the `Copy` function (and is often implemented using it!), but
// `AssignNode` may also try to take faster shortcuts in some implementations, when it detects they're possible.
// (For example, for typed nodes, if they're the same type, lots of checking can be skipped.
// For nodes implemented with pointers, lots of copying can be skipped.
// For nodes that can detect the argument has the same memory layout, faster copy mechanisms can be used; etc.)
//
// Why do both this and the NodeBuilder interface exist?
// In short: NodeBuilder is when you want to cause an allocation;
// NodeAssembler can be used to just "fill in" memory.
// (In the internal gritty details: separate interfaces, one of which lacks a
// `Build` method, helps us write efficient library internals: avoiding the
// requirement to be able to return a Node at any random point in the process
// relieves internals from needing to implement 'freeze' features.
// This is useful in turn because implementing those 'freeze' features in a
// language without first-class/compile-time support for them (as golang is)
// would tend to push complexity and costs to execution time; we'd rather not.)
type NodeAssembler interface {
	BeginMap(sizeHint int64) (MapAssembler, error)
	BeginList(sizeHint int64) (ListAssembler, error)
	AssignNull() error
	AssignBool(bool) error
	AssignInt(int64) error
	AssignFloat(float64) error
	AssignString(string) error
	AssignBytes([]byte) error
	AssignLink(Link) error

	AssignNode(Node) error // if you already have a completely constructed subtree, this method puts the whole thing in place at once.

	// Prototype returns a NodePrototype describing what kind of value we're assembling.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// Using `this.Prototype().NewBuilder()` to produce a new `Node`,
	// then giving that node to `this.AssignNode(n)` should always work.
	// (Note that this is not necessarily an _exclusive_ statement on what
	// sort of values will be accepted by `this.AssignNode(n)`.)
	Prototype() NodePrototype
}

// MapAssembler assembles a map node!  (You guessed it.)
//
// Methods on MapAssembler must be called in a valid order:
// assemble a key, then assemble a value, then loop as long as desired;
// when finished, call 'Finish'.
//
// Incorrect order invocations will panic.
// Calling AssembleKey twice in a row will panic;
// calling AssembleValue before finishing using the NodeAssembler from AssembleKey will panic;
// calling AssembleValue twice in a row will panic;
// etc.
//
// Note that the NodeAssembler yielded from AssembleKey has additional behavior:
// if the node assembled there matches a key already present in the map,
// that assembler will emit the error!
type MapAssembler interface {
	AssembleKey() NodeAssembler   // must be followed by call to AssembleValue.
	AssembleValue() NodeAssembler // must be called immediately after AssembleKey.

	AssembleEntry(k string) (NodeAssembler, error) // shortcut combining AssembleKey and AssembleValue into one step; valid when the key is a string kind.

	Finish() error

	// KeyPrototype returns a NodePrototype that knows how to build keys of a type this map uses.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// For all Data Model maps, this will answer with a basic concept of "string".
	// For Schema typed maps, this may answer with a more complex type
	// (potentially even a struct type or union type -- anything that can have a string representation).
	KeyPrototype() NodePrototype

	// ValuePrototype returns a NodePrototype that knows how to build values this map can contain.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// ValuePrototype requires a parameter describing the key in order to say what
	// NodePrototype will be acceptable as a value for that key, because when using
	// struct types (or union types) from the Schemas system, they behave as maps
	// but have different acceptable types for each field (or member, for unions).
	// For plain maps (that is, not structs or unions masquerading as maps),
	// the empty string can be used as a parameter, and the returned NodePrototype
	// can be assumed applicable for all values.
	// Using an empty string for a struct or union will return nil,
	// as will using any string which isn't a field or member of those types.
	//
	// (Design note: a string is sufficient for the parameter here rather than
	// a full Node, because the only cases where the value types vary are also
	// cases where the keys may not be complex.)
	ValuePrototype(k string) NodePrototype
}

type ListAssembler interface {
	AssembleValue() NodeAssembler

	Finish() error

	// ValuePrototype returns a NodePrototype that knows how to build values this map can contain.
	//
	// You often don't need this (because you should be able to
	// just feed data and check errors), but it's here.
	//
	// ValuePrototype, much like the matching method on the MapAssembler interface,
	// requires a parameter specifying the index in the list in order to say
	// what NodePrototype will be acceptable as a value at that position.
	// For many lists (and *all* lists which operate exclusively at the Data Model level),
	// this will return the same NodePrototype regardless of the value of 'idx';
	// the only time this value will vary is when operating with a Schema,
	// and handling the representation NodeAssembler for a struct type with
	// a representation of a list kind.
	// If you know you are operating in a situation that won't have varying
	// NodePrototypes, it is acceptable to call `ValuePrototype(0)` and use the
	// resulting NodePrototype for all reasoning.
	ValuePrototype(idx int64) NodePrototype
}

type NodeBuilder interface {
	NodeAssembler

	// Build returns the new value after all other assembly has been completed.
	//
	// A method on the NodeAssembler that finishes assembly of the data must
	// be called first (e.g., any of the "Assign*" methods, or "Finish" if
	// the assembly was for a map or a list); that finishing method still has
	// all responsibility for validating the assembled data and returning
	// any errors from that process.
	// (Correspondingly, there is no error return from this method.)
	//
	// Note that building via a representation-level NodePrototype or NodeBuilder
	// returns a node at the type level which implements schema.TypedNode.
	// To obtain the representation-level node, you can do:
	//
	//    // builder is at the representation level, so it returns typed nodes
	//    node := builder.Build().(schema.TypedNode)
	//    reprNode := node.Representation()
	Build() Node

	// Resets the builder.  It can hereafter be used again.
	// Reusing a NodeBuilder can reduce allocations and improve performance.
	//
	// Only call this if you're going to reuse the builder.
	// (Otherwise, it's unnecessary, and may cause an unwanted allocation).
	Reset()
}
