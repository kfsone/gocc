token.Context and $Context
==========================

There are circumstances where you may need to provide stateful context
your parser actions. This can be done through the Parser.Context
field and then accessed in the actions with $Context, where it will be
exposed to you as an `interface{}`.

This example demonstrates using this to track filenames in concurrent
parsers.

See cmd/main.go for a minor example of using context, to execute it:

```
	go run ./cmd
```

Lexer and Token Context fields
------------------------------

In addition to the parser having a Context field, the Lexer also
has one -- to facilitate asts that span multiple files/lexes.

This value is automatically copied into the Pos of every token
produced by a lexer. So for a given `Attrib` argument:

```
	func NewIdentifier(nameAttr Attrib) (*Identifier, error) {
		newToken := nameAttr.(*token.Token)
		name     := string(newToken.Lit)
		if prevToken, exists := identifierTable[name]; exists {
			oldPos  := prevToken.Pos
			newPos  := newToken.Pos
			oldFile := oldPos.Context.(Sourcer).Source()
			newFile := newPos.Context.(Sourcer).Source()
			/* ... */
			return nil, err
		}
		/* ... */
	}
```
