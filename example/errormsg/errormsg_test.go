package errormsg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goccmack/gocc/example/errormsg/errors"
	"github.com/goccmack/gocc/example/errormsg/lexer"
	"github.com/goccmack/gocc/example/errormsg/parser"
	"github.com/goccmack/gocc/example/errormsg/token"

	"github.com/stretchr/testify/assert"
)

func parse(t *testing.T, code string, expectPass bool) {
	t.Helper()

	l := lexer.NewLexer([]byte(code))
	p := parser.NewParser()
	sym, err := p.Parse(l)

	if expectPass {
		if assert.Nil(t, err) {
			assert.NotNil(t, sym)
		}
	} else if assert.Error(t, err) {
		assert.Nil(t, sym)
		t.Log(err.Error())
	}
}

func TestParsedErrors(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		candidates := []string{
			"var abcd = 123;",
			"  var  _  :=  abcd  ;",
			"var a  = 1.23 ;",
			"var x;",
		}
		for _, candidate := range candidates {
			candidate := candidate
			t.Run(candidate, func(t *testing.T) {
				parse(t, candidate, true)
			})
		}
	})

	t.Run("EOF", func(t *testing.T) {
		parse(t, "var a = 1", false) // missing ';'
	})

	t.Run("INVALID", func(t *testing.T) {
		// we never specified \n so it's an unknown symbol.
		parse(t, "var a = 1\n", false) // \n instead of ;
	})

	t.Run("one candidate", func(t *testing.T) {
		// first option is var and only var.
		parse(t, "let a = 1;", false)
	})

	t.Run("two candidates", func(t *testing.T) {
		// second field can be an identifier or _
		parse(t, "var = 1;", false)
	})

	t.Run("three candidates", func(t *testing.T) {
		// third field can be one of '=', ':=' or ';'.
		parse(t, "var _ = ;", false)
	})

	t.Run("four candidates", func(t *testing.T) {
		// 'Default' has four possibilities.
		parse(t, "var xyz = {}", false)
	})

	t.Run("extra tokens", func(t *testing.T) {
		parse(t, "var end = 1; oops", false)
	})
}

func mockToken(tokenType token.Type, lit string, line, col int) *token.Token {
	return &token.Token{Type: tokenType, Lit: []byte(lit), Pos: token.Pos{Line: line, Column: col}}
}

func TestErrors_DescribeExpected(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		assert.Equal(t, "unexpected additional tokens", errors.DescribeExpected([]string{}))
	})
	t.Run("single", func(t *testing.T) {
		assert.Equal(t, "expected TREE", errors.DescribeExpected([]string{"TREE"}))
	})
	t.Run("either", func(t *testing.T) {
		assert.Equal(t, "expected either TREE or ENT", errors.DescribeExpected([]string{"TREE", "ENT"}))
	})
	t.Run("oxford comma", func(t *testing.T) {
		t.Run("list of 3", func(t *testing.T) {
			assert.Equal(t, "expected one of TREE, ENT or i am groot", errors.DescribeExpected([]string{"TREE", "ENT", "i am groot"}))
		})
		t.Run("longer list", func(t *testing.T) {
			assert.Equal(t, "expected one of a, b, c, d, e, f, or g", errors.DescribeExpected([]string{"a", "b", "c", "d", "e", "f", "g"}))
		})
	})
}

func TestErrors_DescribeToken(t *testing.T) {
	t.Run("eof", func(t *testing.T) {
		tok := mockToken(token.EOF, "-not-eof-", 1, 1)
		assert.Equal(t, errors.EOFRepresentation, errors.DescribeToken(tok))
	})
	t.Run("eof", func(t *testing.T) {
		tok := mockToken(token.INVALID, "-not-eof-", 1, 1)
		assert.Equal(t, "unknown/invalid token \"-not-eof-\"", errors.DescribeToken(tok))
	})
	t.Run("eof", func(t *testing.T) {
		tok := mockToken(9001, "-not-eof-", 1, 1)
		assert.Equal(t, "\"-not-eof-\"", errors.DescribeToken(tok))
	})
}

// More direct testing by manually constructing Error objects.
func TestErrors_Error(t *testing.T) {
	// anticipated messages are based on this assumption.
	require.Equal(t, "error", errors.Severity)

	t.Run("custom error", func(t *testing.T) {
		err := &errors.Error{ErrorToken: mockToken(999, "", 6, 7), Err: fmt.Errorf("source on fire")}
		if assert.Error(t, err) {
			assert.Equal(t, "6:7: error: source on fire", err.Error())
		}
	})

	t.Run("no tokens", func(t *testing.T) {
		err := &errors.Error{ErrorToken: mockToken(888, "biscuit", 10, 12), ExpectedTokens: []string{}}
		assert.Equal(t, "10:12: error: unexpected additional tokens; got: \"biscuit\"", err.Error())
	})

	t.Run("unexpected EOF", func(t *testing.T) {
		require.Equal(t, "<EOF>", errors.EOFRepresentation)
		err := &errors.Error{ErrorToken: mockToken(token.EOF, "", 7, 11), ExpectedTokens: []string{"something-else"}}
		assert.Equal(t, "7:11: error: expected something-else; got: <EOF>", err.Error())
	})

	t.Run("nominal error", func(t *testing.T) {
		err := &errors.Error{ErrorToken: mockToken(100, "42", 7, 6), ExpectedTokens: []string{"var", "let", "struct"}}
		assert.Equal(t, "7:6: error: expected one of var, let or struct; got: \"42\"", err.Error())
	})
}
