package compiler

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type compiler struct {
	class      string
	r          *bufio.Reader
	w          io.Writer
	vars       []variable
	labelIndex int
}

type variable struct {
	reg  int
	typ  string
	name string
	idx  int
}

const (
	_ int = iota
	r_static
	r_field
	r_arg
	r_local
)

func NewCompiler(r *bufio.Reader, w io.Writer) compiler {
	c := compiler{
		r: r,
		w: w,
	}
	return c
}

func (cr *compiler) Compile() {
	cr.parseClass()
}

// compiles jack file (without extension) in .vm file, if directory, compiles directory
func CompilePath(path string) {
	inFile, e := os.Open(path)
	if e != nil {
		log.Fatal(e)
	}
	defer inFile.Close()
	outName := strings.TrimSuffix(path, filepath.Ext(path)) + ".vm"
	outFile, e := os.Create(outName)
	if e != nil {
		log.Fatal(e)
	}
	defer outFile.Close()
	cr := NewCompiler(bufio.NewReader(inFile), outFile)
	cr.Compile()
}

func (cr compiler) staticN() int {
	return getcount(cr, r_static)
}

func (cr compiler) fieldN() int {
	return getcount(cr, r_field)
}

func (cr compiler) argN() int {
	return getcount(cr, r_arg)
}

func (cr compiler) localN() int {
	return getcount(cr, r_local)
}

func getcount(c compiler, reg int) int {
	var n int
	for _, v := range c.vars {
		if v.reg == reg {
			n++
		}
	}
	return n
}

func (cr *compiler) addStatic(typ string, name string) {
	addvar(cr, r_static, typ, name)
}

func (cr *compiler) addField(typ string, name string) {
	addvar(cr, r_field, typ, name)
}

func (cr *compiler) addArg(typ string, name string) {
	addvar(cr, r_arg, typ, name)
}

func (cr *compiler) addLocal(typ string, name string) {
	addvar(cr, r_local, typ, name)
}

func addvar(cr *compiler, reg int, typ string, name string) {
	cr.vars = append(cr.vars, variable{
		reg:  reg,
		typ:  typ,
		name: name,
		idx:  getcount(*cr, reg),
	})
}

func (cr *compiler) addVar(mod string, typ string, name string) {
	switch mod {
	case _static:
		addvar(cr, r_static, typ, name)
	case _field:
		addvar(cr, r_field, typ, name)
	case "":
		addvar(cr, r_local, typ, name)
	default:
		fail(cr.r, "unknown var")
	}
}

// return region type and index
func (cr *compiler) getvar(name string) (int, string, int) {
	for _, v := range cr.vars {
		if v.name == name {
			return v.reg, v.typ, v.idx
		}
	}
	return 0, "", 0
}

func (cr *compiler) clearlocals() {
	tmp := cr.vars[:0]
	for _, v := range cr.vars {
		if v.reg != r_local && v.reg != r_arg {
			tmp = append(tmp, v)
		}
	}
	cr.vars = tmp
}

func (cr *compiler) nextLabel(name string) string {
	label := strings.ToUpper(cr.class) + "_" + strings.ToUpper(name) + strconv.Itoa(cr.labelIndex)
	cr.labelIndex++
	return label
}