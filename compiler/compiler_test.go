package compiler

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func testcompiler(input string, w io.Writer) compiler {
	r := bufio.NewReader(strings.NewReader(input))
	cr := newCompiler(r, w)
	cr.class = "Test"
	cr.addStatic(_int, "static")
	cr.addField(_int, "field")
	cr.addArg(_int, "arg")
	cr.addLocal(_int, "local")
	return cr
}

func TestVars(t *testing.T) {
	var c compiler
	c.addStatic(_int, "static_int1")
	c.addStatic(_int, "static_int2")
	c.addField(_int, "field_int1")
	c.addArg(_int, "arg_int1")
	c.addLocal(_int, "local_int1")
	c.addLocal(_int, "local_int2")
	if c.localN()+c.argN()+c.fieldN() != 4 {
		log.Fatal("var table not populated")
	}
	_, _, i := c.getvar("local_int2")
	if i != 1 {
		log.Fatal("var table not populated")
	}
	c.clearlocals()
	if c.localN()+c.argN() != 0 {
		log.Fatal("var table not cleared")
	}
}

func TestParseFileSmall(t *testing.T) {
	testParseFile(t, "test/Seven/Main")
}

func TestParseFileLarge(t *testing.T) {
	testParseFile(t, "test/Pong/PongGame")
}

func testParseFile(t *testing.T, name string) {
	source := name + ".jack"
	sourcefile, e := os.Open(source)
	FAIL(e)
	sourcereader := bufio.NewReader(sourcefile)

	code := name + ".vm"
	codefile, e := os.Open(code)
	FAIL(e)
	codebytes, e := ioutil.ReadAll(codefile)
	FAIL(e)

	buf := bytes.Buffer{}
	cr := newCompiler(sourcereader, &buf)
	cr.Compile()

	if !bytes.Equal(buf.Bytes(), codebytes) {
		t.Fatal("file compilation failed")
	}
}

func FAIL(e error) {
	if e != nil {
		panic(e)
	}
}
