package compiler

import (
	"bufio"
	"strings"
	"testing"
)

func TestSkipToChar(t *testing.T) {
	in :=
		`
// comment

/* some comment
** 
*/
	foo`
	r := bufio.NewReader(strings.NewReader(in))
	if skiptoChar(r) != true {
		t.Fatal("parsing emty space")
	}
	if readliteral(r) != "foo" {
		t.Fatal("parsing literal with spaces")
	}
}

func TestSpaceLiteral(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\r\n\r\n  foo"))
	if peekliteral(r) != "foo" || needliteral(r) != "foo" {
		t.Fatal("parsing literal with spaces")
	}
}

func TestNewlineToken(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\nfoo"))
	if peekliteral(r) != "foo" || needliteral(r) != "foo" {
		t.Fatal("parsing new line token")
	}
}

func TestNewlineChar(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\nfoo"))
	if peekchar(r) != 'f' {
		t.Fatal("parsing new line char")
	}
	needchar(r, 'f')
}

func TestCommentToken(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("//some comment\nfoo"))
	if peekliteral(r) != "foo" || needliteral(r) != "foo" {
		t.Fatal("parsing new line token")
	}
}

func TestCommentChar(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("//some comment\nfoo"))
	if peekchar(r) != 'f' {
		t.Fatal("parsing new line char")
	}
	needchar(r, 'f')
}
