package compiler

import (
	"bytes"
	"testing"
)

// EXPRESSIONS

func TestIntExpPriority01(t *testing.T) {
	exp := expression{
		intTerm{3},
		'*',
		expression{
			intTerm{5},
			'+',
			intTerm{7},
		},
	}
	expected := `push constant 3
push constant 5
push constant 7
add
call Math.multiply 2
`
	testExpStruct(t, exp, expected)
}

func TestIntExpPriority02(t *testing.T) {
	exp := expression{
		expression{
			intTerm{5},
			'+',
			intTerm{7},
		},
		'*',
		intTerm{3},
	}
	expected := `push constant 5
push constant 7
add
push constant 3
call Math.multiply 2
`
	testExpStruct(t, exp, expected)
}

func TestExpression(t *testing.T) {
	expected := `push constant 2
push constant 2
push constant 3
call Math.multiply 2
push constant 4
call Math.multiply 2
add
push constant 1
sub
push constant 2
push constant 1
call Math.divide 2
sub
`
	testExpString(t, "2+2*3*4-1-2/1", expected)
}

func TestExpressionBrackets(t *testing.T) {
	expected := `push constant 2
push constant 3
add
push constant 5
call Math.multiply 2
`
	testExpString(t, "(2+3)*5", expected)

}

func TestExpressionBracketsComplex(t *testing.T) {
	expected :=
		`push constant 2
push constant 3
push constant 2
sub
push constant 4
add
call Math.multiply 2
push constant 2
push constant 3
sub
call Math.divide 2
`
	testExpString(t, "(2*(3-2+4)/(2-3))", expected)
}

func TestExpressionVar(t *testing.T) {
	expected :=
		`push constant 2
push constant 10
lt
`
	testExpString(t, "2 < 10", expected)
}

func TestExpressionLogicalSimple(t *testing.T) {
	expected :=
		`push constant 1
neg
push constant 10
push constant 0
gt
and
`
	testExpString(t, "true&(10>0)", expected)
}

func testExpStruct(t *testing.T, exp expression, expected string) {
	buf := bytes.Buffer{}
	cr := testcompiler("", &buf)
	cr.code(exp)
	if buf.String() != expected {
		t.Fatalf("error in code generation\nRESULT\n%s\nEXPECTING\n%s\n", buf.String(), expected)
	}
}

func testExpString(t *testing.T, exp string, expected string) {
	buf := bytes.Buffer{}
	cr := testcompiler(exp, &buf)
	parsed := cr.parseExpr()
	cr.code(parsed)
	if buf.String() != expected {
		t.Fatalf("error in code generation\nRESULT\n%s\nEXPECTING\n%s\n", buf.String(), expected)
	}
}


// STATEMENTS

func TestParseClass(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("class Foo { field int bar, baz; static int baq; method void Bar() {} }", &buf)
	cr.parseClass()
}
//
func TestParseFnBody(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("{ var int foo, bar; }", &buf)
	cr.parseFnBody()
}

func TestParseLet(t *testing.T) {
	cr := testcompiler("let foo = bar;", &bytes.Buffer{})
	cr.parseLetStmt()
}

func TestLet(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("let static=1;", &buf)
	expected := "push constant 1\npop static 0\n"
	stmt := cr.parseLetStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestParseDo(t *testing.T) {
	cr := testcompiler("do foo();", &bytes.Buffer{})
	cr.parseDoStmt()
}

func TestParseReturn(t *testing.T) {
	cr := testcompiler("return 2+2;", &bytes.Buffer{})
	cr.parseReturnStmt()
}

func TestReturn(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("return;", &buf)
	expected := "push constant 0\nreturn\n"
	stmt := cr.parseReturnStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestParseWhile(t *testing.T) {
	cr := testcompiler("while (true) { } ;", &bytes.Buffer{})
	cr.parseWhileStmt()
}

func TestWhile(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("while (true) {}", &buf)
	expected := "label TEST_WHILE_START0\npush constant 1\nneg\npush constant 0\neq\nif-goto TEST_WHILE_END1\ngoto TEST_WHILE_START0\nlabel TEST_WHILE_END1\n"
	stmt := cr.parseWhileStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestParseIf(t *testing.T) {
	cr := testcompiler("if (false) {} else {}", &bytes.Buffer{})
	cr.parseIfStmt()
}

func TestCallStatic(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("do Foo.bar();", &buf)
	expected := "call Foo.bar 0\npop temp 0\n"
	stmt := cr.parseDoStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestCallMethod(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("do bar();", &buf)
	expected := "push pointer 0\ncall Test.bar 1\npop temp 0\n"
	stmt := cr.parseDoStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestIf(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("if (false) {} else {}", &buf)
	expected := "push constant 0\npush constant 0\neq\nif-goto TEST_IF_ELSE0\ngoto TEST_IF_END1\nlabel TEST_IF_ELSE0\nlabel TEST_IF_END1\n"
	stmt := cr.parseIfStmt()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestSubroutineMethod(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("poke(8000 + 1, 1)", &buf)
	expected :=
		`push pointer 0
push constant 8000
push constant 1
add
push constant 1
call Test.poke 3
`
	stmt := cr.parseTerm()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}

func TestSubroutineStatic(t *testing.T) {
	buf := bytes.Buffer{}
	cr := testcompiler("Memory.poke(8000 + 1, 1)", &buf)
	expected :=
		`push constant 8000
push constant 1
add
push constant 1
call Memory.poke 2
`
	stmt := cr.parseTerm()
	cr.code(stmt)
	if buf.String() != expected {
		println(buf.String())
		t.Fail()
	}
}
