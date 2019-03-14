package compiler

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	_true        = "true"
	_false       = "false"
	_null        = "null"
	_this        = "this"
	_static      = "static"
	_field       = "field"
	_constructor = "constructor"
	_function    = "function"
	_method      = "method"
	_let         = "let"
	_if          = "if"
	_else        = "else"
	_while       = "while"
	_do          = "do"
	_return      = "return"
	_var         = "var"
	_int         = "int"
	_char        = "char"
	_boolean     = "boolean"
)

func (cr *compiler) code(token interface{}) { // fixme double cases
	switch t := token.(type) {
	case ifStmtToken:
		ifelse := cr.nextLabel("if_else")
		ifend := cr.nextLabel("if_end")
		cr.code(t.cond)
		cr.pushConst(0)
		cr.line("eq")
		cr.line("if-goto " + ifelse)
		for _, stmt := range t.ifst {
			cr.code(stmt)
		}
		cr.line("goto " + ifend)
		cr.line("label " + ifelse)
		for _, stmt := range t.elsest {
			cr.code(stmt)
		}
		cr.line("label " + ifend)
	case letStmtToken:
		cr.code(t.exp)
		cr.popVar(cr, t.name)
	case whileStmtToken:
		start := cr.nextLabel("while_start")
		end := cr.nextLabel("while_end")
		cr.line("label " + start)
		cr.code(t.cond)
		cr.pushConst(0)
		cr.line("eq")
		cr.line("if-goto " + end)
		for _, stmt := range t.stmts {
			cr.code(stmt)
		}
		cr.line("goto " + start)
		cr.line("label " + end)
	case doStmtToken:
		cr.code(t.stmt)
		cr.popTemp(0) // discard result
	case returnStmtToken:
		if t.exp.left == nil {
			cr.pushConst(0)
		} else {
			cr.code(t.exp)
		}
		cr.line("return")
	case fnToken:
		cr.clearlocals()
		if t.mod == _method {
			cr.addArg(cr.class, "this")
		}
		if t.mod == _constructor {
			cr.addLocal(cr.class, "this")
		}
		// first populate var tables
		for _, param := range t.params {
			cr.code(param)
		}
		for _, v := range t.body.vars {
			cr.code(v)
		}
		cr.linef("function %s.%s %d", cr.class, t.name, cr.localN())
		if t.mod == _method {
			cr.pushArg(0)
			cr.line("pop pointer 0")
		}
		if t.mod == _constructor {
			cr.pushConst(cr.fieldN())
			cr.linef("call Memory.alloc %d", 1)
			cr.popLocal(0) // this
			cr.pushLocal(0)
			cr.line("pop pointer 0")
		}
		for _, v := range t.body.stmts {
			cr.code(v)
		}
	case varDeclToken:
		for _, name := range t.names {
			cr.addVar(t.mod, t.typ, name)
		}
	case paramListToken:
		for i := range t.name {
			cr.addArg(t.typ[i], t.name[i])
		}
	case subroutineTerm:
		class := t.name
		argsN := len(t.exprs)

		if t.local {
			cr.pushPointer(0)
			argsN++
		}
		reg, typ, _ := cr.getvar(t.name)
		if reg != 0 {
			// calling var
			class = typ
			cr.pushVar(cr, t.name)
			argsN++
		}
		for _, exp := range t.exprs {
			cr.code(exp)
		}
		cr.linef("call %s.%s %d", class, t.fn, argsN)
	case expression:
		if t.left == nil {
			break
		}
		cr.code(t.left)
		if t.right == nil {
			break
		}
		cr.code(t.right)
		switch t.op {
		case '+':
			cr.line("add")
		case '-':
			cr.line("sub")
		case '*':
			cr.line("call Math.multiply 2")
		case '/':
			cr.line("call Math.divide 2")
		case '&':
			cr.line("and")
		case '|':
			cr.line("or")
		case '<':
			cr.line("lt")
		case '>':
			cr.line("gt")
		case '=':
			cr.line("eq")
		default:
			fail(cr.r, "unsupported expression operation %c", t.op)
		}
	case intTerm:
		cr.pushConst(t.int)
	case strTerm:
		cr.pushConst(len(t.string))
		cr.line("call String.new 1")
		for _, c := range t.string {
			cr.pushConst(int(c))
			cr.line("call String.appendChar 2")
		}
	case keywordTerm:
		switch t.string {
		case _true:
			cr.pushConst(-1)
		case _false, _null:
			cr.pushConst(0)
		case _this:
			cr.pushPointer(0)
		default:
			fail(cr.r, "wrong keyword token %s", t.string)
		}
	case varTerm:
		cr.pushVar(cr, t.string)
	case unaryOpTerm:
		cr.code(t.term)
		switch t.byte {
		case '-':
			cr.line("neg")
		case '~':
			cr.line("not")
		default:
			fail(cr.r, "wrong unary op %c", t.byte)
		}

	default:
		fail(cr.r, "unknown token '%s' %v", reflect.TypeOf(token), t)
	}
}

func (cr compiler) line(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	io.WriteString(cr.w, s+"\n")
}
func (cr compiler) linef(s ...interface{}) {
	if len(s) < 2 {
		panic(0)
	}
	format := s[0].(string)
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	io.WriteString(cr.w, fmt.Sprintf(format, s[1:]...))
}
func w(s ...interface{}) string {
	format := s[0].(string) + "\n"
	if len(s) == 1 {
		return format
	}
	return fmt.Sprintf(format, s[1:]...)
}
func (cr *compiler) popLocal(i int) {
	cr.linef("pop local %d", i)
}
func (cr *compiler) popTemp(i int) {
	cr.linef("pop temp %d", i)
}
func (cr *compiler) popThis(i int) {
	cr.linef("pop this %d", i) // fixme
}
func (cr *compiler) popThat(i int) {
	cr.linef("pop that %d", i) // fixme
}
func (cr *compiler) popStatic(i int) {
	cr.linef("pop static %d", i)
}
func (cr *compiler) popArg(i int) {
	cr.linef("pop argument %d", i)
}

func (cr *compiler) pushLocal(i int) {
	cr.linef("push local %d", i)
}
func (cr *compiler) pushTemp(i int) {
	cr.linef("push temp %d", i)
}
func (cr *compiler) pushThis(i int) {
	cr.linef("push this %d", i) // fixme
}
func (cr *compiler) pushThat(i int) {
	cr.linef("push that %d", i) // fixme
}
func (cr *compiler) pushStatic(i int) {
	cr.linef("push static %d", i)
}
func (cr *compiler) pushArg(i int) {
	cr.linef("push argument %d", i)
}
func (cr *compiler) pushConst(i int) {
	var s string
	if i < 0 {
		s = w("push constant %d", -i)
		s += w("neg")
	} else {
		s = w("push constant %d", i)
	}
	cr.line(s)
}
func (cr *compiler) pushPointer(i int) {
	cr.linef("push pointer %d", i)
}
func (cr *compiler) pushVar(comp *compiler, name string) {
	reg, _, i := comp.getvar(name)
	switch reg {
	case regLocal:
		cr.pushLocal(i)
	case regArg:
		cr.pushArg(i)
	case regField:
		cr.pushThis(i)
	case regStatic:
		cr.pushStatic(i)
	default:
		fail(cr.r, "var undefined %s", name)
	}
}
func (cr *compiler) popVar(comp *compiler, name string) {
	reg, _, i := comp.getvar(name)
	switch reg {
	case regLocal:
		cr.popLocal(i)
	case regArg:
		cr.popArg(i)
	case regField:
		cr.popThis(i)
	case regStatic:
		cr.popStatic(i)
	default:
		fail(cr.r, "var undefined %s", name)
	}
}
