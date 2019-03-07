package compiler

/*
need* and peek* stmts skip space to nearest token, toleratingT comments
read* reads the nearest (no comments)
*/

import (
	"bufio"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
)

const PEEKBUFFER = 80

func readint(r *bufio.Reader) int {
	bytes := []byte{}
	for {
		c, _ := r.ReadByte()
		if !isInteger(c) {
			r.UnreadByte()
			i, _ := strconv.Atoi(string(bytes))
			return i
		}
		bytes = append(bytes, c)
	}
}

func readliteral(r *bufio.Reader) string {
	bytes := []byte{}
	for {
		c, _ := r.ReadByte()
		if !isLiteral(c) {
			r.UnreadByte()
			return string(bytes)
		}
		bytes = append(bytes, c)
	}
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isLiteral(c byte) bool {
	return 'A' <= c && c <= 'z'
}

func isInteger(c byte) bool {
	return '0' <= c && c <= '9'
}

// true if space was skipped
func skiptoChar(r *bufio.Reader) bool {
	var space bool

	for {
		b, _ := r.Peek(2)
		if len(b) == 0 {
			return space
		}
		if len(b) == 1 {
			if isSpace(b[0]) {
				r.ReadByte()
				space = true
			}
			return space
		}
		if isSpace(b[0]) {
			r.ReadByte()
			space = true
			continue
		}
		if b[0] == '/' && b[1] == '/' {
			r.ReadString('\n')
			space = true
			continue
		}
		if b[0] == '/' && b[1] == '*' {
			skipMultilineComment(r)
			space = true
			continue
		}
		return space
	}
}

func skipMultilineComment(r *bufio.Reader) {
	// skip opening tag
	r.ReadByte()
	r.ReadByte()
	var expectEnd bool
	for {
		c, _ := r.ReadByte()
		switch c {
		case '*':
			expectEnd = true
		case '/':
			if expectEnd {
				return
			}
		case 0:
			return
		default:
			expectEnd = false
		}
	}
}

func needliteral(r *bufio.Reader) string {
	skiptoChar(r)
	s := readliteral(r)
	if s == "" {
		fail(r, "expecting literal")
	}
	return s
}

// skipping space
func needchar(r *bufio.Reader, b byte) {
	skiptoChar(r)
	c, _ := r.ReadByte()
	if c == b {
		return
	}
	fail(r, "expecting symbol '%c'", b)
}

func peekchar(r *bufio.Reader) byte {
	skiptoChar(r)
	c, _ := r.Peek(1)
	if len(c) == 0 {
		return 0
	}
	return c[0]
}

func peekliteral(r *bufio.Reader) string {
	skiptoChar(r)
	bytes := []byte{}
	peek, _ := r.Peek(PEEKBUFFER)
	if len(peek) == 0 {
		return ""
	}
	for _, c := range peek {
		if !isLiteral(c) {
			break
		}
		bytes = append(bytes, c)
	}
	return string(bytes)
}

func fail(r *bufio.Reader, args ...interface{}) {
	r.UnreadByte()
	buf, _, e := r.ReadLine()
	context := string(buf)
	if context == "" && e != nil {
		context = e.Error()
	}
	debug.PrintStack()
	format := args[0].(string)
	if len(args) == 1 {
		log.Fatalf("%s at: '%s'\n", format, context)
	} else {
		msg := fmt.Sprintf(format, args[1:]...)
		log.Fatalf("%s at: '%s'\n", msg, context)
	}
}
