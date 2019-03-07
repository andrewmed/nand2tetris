package compiler

import "log"

// TOP LEVEL TOKENS (produce output)

func (cr *compiler) parseClass() {
	if needliteral(cr.r) != "class" {
		fail(cr.r, "expecting classfile")
	}
	cr.class = needliteral(cr.r)
	needchar(cr.r, '{')
	var literal string
	for literal = peekliteral(cr.r); literal == _static || literal == _field; literal = peekliteral(cr.r) {
		cr.parseClassVar()
	}
	for ; literal == _constructor || literal == _function || literal == _method; literal = peekliteral(cr.r) {
		cr.parseFn()
	}
	needchar(cr.r, '}')
}

func (cr *compiler) parseClassVar()  {
	var token varDeclToken
	token.mod = needliteral(cr.r)
	token.typ = needliteral(cr.r)
	token.names = append(token.names, needliteral(cr.r))
	for peekchar(cr.r) == ',' {
		needchar(cr.r, ',')
		token.names = append(token.names, needliteral(cr.r))
	}
	needchar(cr.r, ';')
	cr.code(token)
}

func (cr *compiler) parseFn()  {
	var token fnToken
	token.mod = needliteral(cr.r)
	token.rettype = needliteral(cr.r)
	token.name = needliteral(cr.r)
	needchar(cr.r, '(')
	if peekchar(cr.r) != ')' {
		for {
			token.params = append(token.params, cr.parseParamList())
			if peekchar(cr.r) == ')' {
				break
			}
			needchar(cr.r, ',')
		}
	}
	needchar(cr.r, ')')
	token.body = cr.parseFnBody()
	cr.code(token)
}

// LOWER TOKENS (return structs)

func (cr *compiler) parseTerm() interface{} {
	skiptoChar(cr.r)
	c, _ := cr.r.ReadByte()
	if isLiteral(c) {
		cr.r.UnreadByte()
		return cr.parseLiteralTerm()
	}
	if isInteger(c) {
		cr.r.UnreadByte()
		return intTerm{readint(cr.r)}
	}
	switch c {
	case '(':
		term := cr.parseExpr()
		needchar(cr.r, ')')
		return term
	case '"':
		bytes, _ := cr.r.ReadBytes('"')
		s := string(bytes[:len(bytes) - 2])
		term := strTerm{s}
		return term
	case '-', '~':
		term := unaryOpTerm{
			byte: c,
			term: cr.parseTerm(),
		}
		return term
	default:
		cr.r.UnreadByte()
	}
	return nil
}

func (cr *compiler) parseLiteralTerm() interface{} {
	s := readliteral(cr.r)
	if s == _true || s == _false || s == _null || s == _this {
		return keywordTerm{s}
	} else {
		switch peekchar(cr.r) {
		case '.':
			// subroutine call
			needchar(cr.r, '.')
			fn := readliteral(cr.r)
			needchar(cr.r, '(')
			exprs := cr.parseExprList()
			needchar(cr.r, ')')
			return subroutineTerm{
				name:  s,
				fn:    fn,
				exprs: exprs,
			}
		case '(':
			// subroutine call
			needchar(cr.r, '(')
			exprs := cr.parseExprList()
			needchar(cr.r, ')')
			return subroutineTerm{
				local: true,
				name:  cr.class,
				fn:    s,
				exprs: exprs,
			}
		default:
			return varTerm{s}
		}
	}
}

func (cr *compiler) parseOp() byte {
	if cr == nil {
		return 0
	}
	skiptoChar(cr.r)
	c, _ := cr.r.ReadByte()
	switch c {
	case '+', '-', '*', '/', '&', '|', '<', '>', '=':
		return c
	default:
		cr.r.UnreadByte()
		return 0
	}
}

func (cr *compiler) parseExprList() []expression {
	var exprs []expression
	if cr == nil {
		return exprs
	}
	exp := cr.parseExpr()
	if exp.left == nil {
		return exprs
	}
	exprs = append(exprs, exp)
	for peekchar(cr.r) == ',' {
		needchar(cr.r, ',')
		exprs = append(exprs, cr.parseExpr())
	}
	return exprs
}

func (cr *compiler) parseExpr() expression {
	var root expression
	if cr == nil {
		return root
	}
	tree := tree{}
	t := cr.parseTerm()
	if t == nil {
		return expression{}
	}
	tree.terms = append(tree.terms, t)
	for {
		op := cr.parseOp()
		if op == 0 {
			break
		}
		tree.ops = append(tree.ops, op)
		t = cr.parseTerm()
		if t == nil {
			fail(cr.r, "incomplete expression")
		}
		tree.terms = append(tree.terms, t)
	}
	root = nextE(&tree, 0)
	for {
		exp := nextE(&tree, 0)
		if exp.left == nil {
			return root
		}
		if len(tree.ops) == 0 {
			root.right = exp
			return root
		}
		root = expression{
			left:  root,
			op:    tree.ops[0],
			right: exp,
		}
		tree.ops = tree.ops[1:]
	}
}

func nextE(tree *tree, priority int) expression {
	var exp expression
	if len(tree.terms) == 0 {
		return exp
	}
	exp.left = tree.terms[0]
	tree.terms = tree.terms[1:]
	for len(tree.ops) > 0 && prio(tree.ops[0]) > priority {
		op := tree.ops[0]
		tree.ops = tree.ops[1:]
		exp = expression{
			left:  exp,
			op:    op,
			right: nextE(tree, prio(op)),
		}
	}
	return exp
}

func (cr *compiler) parseReturnStmt() returnStmtToken {
	var token returnStmtToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _return {
		fail(cr.r, "expecting return stmt")
	}
	token.exp = cr.parseExpr()
	needchar(cr.r, ';')
	return token
}

func (cr *compiler) parseDoStmt() doStmtToken {
	var token doStmtToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _do {
		fail(cr.r, "expecting do stmt")
	}
	token.stmt = cr.parseTerm()
	needchar(cr.r, ';')
	return token

}

func (cr *compiler) parseWhileStmt() whileStmtToken {
	var token whileStmtToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _while {
		fail(cr.r, "expecting while stmt")
	}
	needchar(cr.r, '(')
	token.cond = cr.parseExpr()
	needchar(cr.r, ')')
	needchar(cr.r, '{')
	token.stmts = cr.parseStmts(peekliteral(cr.r))
	needchar(cr.r, '}')
	return token
}

func (cr *compiler) parseIfStmt() ifStmtToken {
	var token ifStmtToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _if {
		fail(cr.r, "expecting if stmt")
	}
	needchar(cr.r, '(')
	token.cond = cr.parseExpr()
	needchar(cr.r, ')')
	needchar(cr.r, '{')
	token.ifst = cr.parseStmts(peekliteral(cr.r))
	needchar(cr.r, '}')
	if peekliteral(cr.r) == _else {
		needliteral(cr.r)
		needchar(cr.r, '{')
		token.elsest = cr.parseStmts(peekliteral(cr.r))
		needchar(cr.r, '}')
	}
	return token
}

func (cr *compiler) parseLetStmt() letStmtToken {
	var token letStmtToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _let {
		fail(cr.r, "expecting let stmt")
	}
	token.name = needliteral(cr.r)
	needchar(cr.r, '=')
	token.exp = cr.parseExpr()
	needchar(cr.r, ';')
	return token
}

func (cr *compiler) parseFnBody() fnBodyToken {
	var token fnBodyToken
	if cr == nil {
		return token
	}
	needchar(cr.r, '{')
	peek := peekliteral(cr.r)
	for {
		if peek != "var" {
			break
		}
		token.vars = append(token.vars, cr.parseFnVar())
		peek = peekliteral(cr.r)
	}
	token.stmts = cr.parseStmts(peek)
	needchar(cr.r, '}')
	return token
}

// needs peeked token and peeks at the end
func (cr *compiler) parseStmts(peek string) []interface{} {
	var t []interface{}
	if cr == nil {
		return t
	}
	for {
		switch peek {
		case _let:
			t = append(t, cr.parseLetStmt())
		case _if:
			t = append(t, cr.parseIfStmt())
		case _while:
			t = append(t, cr.parseWhileStmt())
		case _do:
			t = append(t, cr.parseDoStmt())
		case _return:
			t = append(t, cr.parseReturnStmt())
		default:
			return t
		}
		peek = peekliteral(cr.r)
	}
}

func (cr *compiler) parseFnVar() varDeclToken {
	var token varDeclToken
	if cr == nil {
		return token
	}
	if needliteral(cr.r) != _var {
		fail(cr.r, "expecting var declaration")
	}
	token.typ = needliteral(cr.r)
	token.names = append(token.names, needliteral(cr.r))
	for peekchar(cr.r) == ',' {
		needchar(cr.r, ',')
		token.names = append(token.names, needliteral(cr.r))
	}
	needchar(cr.r, ';')
	return token
}

func (cr *compiler) parseParamList() paramListToken {
	var token paramListToken
	if cr == nil {
		return token
	}
	token.typ = append(token.typ, needliteral(cr.r))
	token.name = append(token.name, needliteral(cr.r))
	return token
}

func prio(op byte) int {
	switch op {
	case '&', '|', '<', '>', '=':
		return 5
	case '+', '-':
		return 10
	case '*', '/':
		return 20
	default:
		log.Fatalf("unknown operation '%c'", op)
	}
	return -1 // need to compile
}

