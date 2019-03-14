package compiler

// TOKENS

// 'class' className '{' classVarDec* subroutineDec* '}'

// ('static' | 'field') type varName (',' varName)* ';'
type varDeclToken struct {
	mod   string
	typ   string
	names []string
}

// ('constructor' | 'function' | 'method') ('void' | type) subroutineName '(' parameterList ')' subroutineBody
type fnToken struct {
	mod     string
	rettype string
	name    string
	params  []paramListToken
	body    fnBodyToken
}

// 	( (type varName) (',' type varName)* )?
type paramListToken struct {
	typ  []string
	name []string
}

// '{' varDec* statements '}'
type fnBodyToken struct {
	vars  []varDeclToken
	stmts []interface{} // stmt
}

// 'let' varName ('[' expression ']')? '=' expression ';'
type letStmtToken struct {
	name string
	exp  expression
}

// 'if' '(' expression ')' '{' statements '}' ('else' '{' statements '}')?
type ifStmtToken struct {
	cond   expression
	ifst   []interface{} // stmt
	elsest []interface{} // stmt
}

// 'while' '(' expression ')' '{' statements '}'
type whileStmtToken struct {
	cond  expression
	stmts []interface{} // stmt
}

// 'do' subroutineCall ';'
type doStmtToken struct {
	stmt interface{} // stmt
}

// 'return' ( expression )? ';'
type returnStmtToken struct {
	exp expression
}

// TERMS

type expression struct {
	left  interface{} // term
	op    byte
	right interface{} // term
}

type subroutineTerm struct {
	local bool
	name  string // classname or var name
	fn    string
	exprs []expression
}

type tree struct {
	terms []interface{} // term
	ops   []byte
}

type intTerm struct {
	int
}

type strTerm struct {
	string
}

type keywordTerm struct {
	string
}

type varTerm struct {
	string
}

type unaryOpTerm struct {
	byte             // operation - ~
	term interface{} // term
}
