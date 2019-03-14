// Package vmtranslator translates (VM) code to assembly (ASM), single pass
package vmtranslator

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const comments = true

type cmdType int

const (
	cmdArithmetic cmdType = iota
	cmdPush
	cmdPop
	cmdLabel
	cmdGoto
	cmdIf
	cmdFunction
	cmdReturn
	cmdCall
)

const (
	sp       = "sp"
	local    = "local"
	argument = "argument"
	this     = "this"
	that     = "that"
	temp     = "temp"
)

// pointers to mem indexes
var ptr = map[string]int{
	sp:       0, // stack pointer
	local:    1, // LCL
	argument: 2, // ARG
	this:     3,
	that:     4,
	temp:     5,
}

type cmdStruct struct {
	cmd  cmdType
	arg1 string
	arg2 int
}

// VMTranslator translates VM to ASM
type VMTranslator struct {
	jumpIndex int // jump label index
	strings.Builder
	name string
}

// Bootstrap adds init section
func Bootstrap(b *VMTranslator) {
	log.Println("Compiling bootstrap code")
	// set sp
	if comments {
		b.c("// bootstrap section")
	}
	b.a(256)
	b.c("D=A")
	b.a(sp)
	b.c("M=D")
	//// then we zero local, arg, this and that
	b.a(local)
	b.c("M=0")
	b.a(argument)
	b.c("M=0")
	b.a(this)
	b.c("M=0")
	b.a(that)
	b.c("M=0")

	// push 5 elements to stack
	b.c("D=0")
	b.pushD()
	b.pushD()
	b.pushD()
	b.pushD()
	b.pushD()
	b.a("Sys.init")
	b.c("0;JMP")
}

// TranslateFile translates one file, returns N of lines processed and ok
func TranslateFile(b *VMTranslator, filename string) (int, bool) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, false
	}
	scanner := bufio.NewScanner(file)
	b.name = strings.TrimSuffix(path.Base(filename), ".vm")
	return translate(b, scanner)
}

func translate(b *VMTranslator, scanner *bufio.Scanner) (int, bool) {
	var errFound bool
	var line int

	for scanner.Scan() {
		line++
		s := scanner.Text()
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "//") {
			if comments {
				b.c(s)
			}
			continue
		}
		if comments {
			b.c("// " + s)
		}
		c, e := parse(s)
		if e != nil {
			log.Printf("Line %d %s: '%s'\n", line, e, s)
			errFound = true
			continue
		}
		if e := generate(b, c); e != nil {
			log.Printf("Line %d %s: '%s'\n", line, e, s)
			errFound = true
			continue
		}
	}
	log.Printf("%d lines processed.\n", line)
	return line, !errFound

}

func parse(s string) (cmdStruct, error) {
	tokens := strings.Split(s, " ")

	switch tokens[0] {
	case "function":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: cmdFunction, arg1: tokens[1], arg2: arg2}, nil
	case "call":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: cmdCall, arg1: tokens[1], arg2: arg2}, nil
	case "goto":
		return cmdStruct{cmd: cmdGoto, arg1: tokens[1]}, nil
	case "if-goto":
		return cmdStruct{cmd: cmdIf, arg1: tokens[1]}, nil
	case "return":
		return cmdStruct{cmd: cmdReturn}, nil
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return cmdStruct{cmd: cmdArithmetic, arg1: tokens[0]}, nil
	case "push":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: cmdPush, arg1: tokens[1], arg2: arg2}, nil
	case "pop":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: cmdPop, arg1: tokens[1], arg2: arg2}, nil
	case "label":
		return cmdStruct{cmd: cmdLabel, arg1: tokens[1]}, nil
	}

	return cmdStruct{}, errors.New("parsing")
}

// A command
// if arg is int adds "@arg"
// if str arg is in  ptr{} add it as pointer val
// otherwise just add as it is (var ref)
func (b *VMTranslator) a(x interface{}) (int, error) {
	switch v := x.(type) {
	case int:
		return b.WriteString(fmt.Sprintf("@%d\n", v))
	case string:
		i, ok := ptr[v]
		if !ok {
			return b.WriteString(fmt.Sprintf("@%s\n", v))
		}
		return b.WriteString(fmt.Sprintf("@%d\n", i))
	default:
		panic("wrong b.a() usage")
	}
}

// B command
// with one arg just adds
// with more formats in Sprintf
func (b *VMTranslator) c(s ...interface{}) (int, error) {
	format := s[0].(string) + "\n"
	if len(s) == 1 {
		return b.WriteString(format)
	}

	return b.WriteString(fmt.Sprintf(format, s[1:]...))
}

// if no arg, returns next jump label
// otherwise, outputs asm label
func (b *VMTranslator) l(s ...string) (string, error) {
	if len(s) == 0 {
		b.jumpIndex++
		return "JUMP" + strconv.Itoa(b.jumpIndex), nil
	}
	_, e := b.WriteString("(" + s[0] + ")\n")
	return "", e
}

// pop from stack to M register
func (b *VMTranslator) popM() (int, error) {
	return b.WriteString("@0\nM=M-1\nA=M\n")
}

// push D register on stack
func (b *VMTranslator) pushD() (int, error) {
	return b.WriteString("@0\nA=M\nM=D\n@0\nM=M+1\n")
}

// pop from stack to D register
func (b *VMTranslator) popD() (int, error) {
	return b.WriteString("@0\nM=M-1\nA=M\nD=M\n")
}

func generate(b *VMTranslator, c cmdStruct) error {
	switch c.cmd {
	case cmdCall:
		// push ret address (see below)
		label, _ := b.l()
		b.a(label)
		b.c("D=A")
		b.pushD()
		// push LCL
		b.a(local)
		b.c("D=M")
		b.pushD()
		// push ARG
		b.a(argument)
		b.c("D=M")
		b.pushD()
		// push THIS
		b.a(this)
		b.c("D=M")
		b.pushD()
		// push ARG
		b.a(argument)
		b.c("D=M")
		b.pushD()
		// ARG = sp - (n of args) - 5
		b.a(sp)
		b.c("D=M")
		b.a(5)
		b.c("D=D-A")
		b.a(c.arg2)
		b.c("D=D-A")
		b.a(argument)
		b.c("M=D")
		// LCL= sp
		b.a(sp)
		b.c("D=M")
		b.a(local)
		b.c("M=D")
		// goto f
		b.a(c.arg1)
		b.c("0;JMP")
		// ret label
		b.l(label)

	case cmdReturn:
		// FRAME = LCL (temp0)
		b.a(local)
		b.c("D=M")
		b.a(temp)
		b.c("M=D")
		// RET = *(FRAME-5) (temp1)
		b.a(5)
		b.c("D=D-A")
		b.a(temp)
		b.c("A=A+1")
		b.c("M=D")
		// ARG 0 = pop()
		b.popD()
		b.a(argument)
		b.c("A=M")
		b.c("M=D")
		// sp = *ARG + 1
		b.c("A=A+1")
		b.c("D=A")
		b.a(sp)
		b.c("M=D")
		// THAT = * (FRAME - 1)
		b.a(temp)
		b.c("D=M")
		b.a(1)
		b.c("A=D-A")
		b.c("D=M")
		b.a(that)
		b.c("M=D")
		// THIS = *( FRAME - 2)
		b.a(temp)
		b.c("D=M")
		b.a(2)
		b.c("A=D-A")
		b.c("D=M")
		b.a(this)
		b.c("M=D")
		// ARG = *(FRAME - 3)
		b.a(temp)
		b.c("D=M")
		b.a(3)
		b.c("A=D-A")
		b.c("D=M")
		b.a(argument)
		b.c("M=D")
		// LCL = *(FRAME - 4)
		b.a(temp)
		b.c("D=M")
		b.a(4)
		b.c("A=D-A")
		b.c("D=M")
		b.a(local)
		b.c("M=D")
		// goto RET
		b.a(temp)
		b.c("A=A+1")
		b.c("A=M")
		b.c("A=M")
		b.c("0;JMP")
	case cmdIf:
		b.popD()
		b.a(c.arg1)
		b.c("D;JNE")
	case cmdGoto:
		b.a(c.arg1)
		b.c("0;JMP")
	case cmdLabel:
		b.c("(%s)", c.arg1)
	case cmdFunction:
		b.c("(%s)", c.arg1)
		for ; c.arg2 > 0; c.arg2-- { // preallocate locals
			b.pushD() // no need to clear, this should be job of c higher level language
		}
	case cmdPop: // pop from stack
		// to receiver memory region
		switch c.arg1 {
		case "this", "that", "local", "argument":
			b.a(c.arg1)
			b.c("D=M") // diff
			b.a(c.arg2)
			b.c("D=D+A")
			b.a(13)
			b.c("M=D")
			b.popD()
			b.a(13)
			b.c("A=M")
			b.c("M=D")
		case "temp":
			b.a(c.arg1)
			b.c("D=A") // diff
			b.a(c.arg2)
			b.c("D=D+A")
			b.a(13)
			b.c("M=D")
			b.popD()
			b.a(13)
			b.c("A=M")
			b.c("M=D")
		case "static":
			b.popD()
			b.c("@%s.%d", b.name, c.arg2)
			b.c("M=D")
		case "pointer": // pointer 0 -> this, pointer 1 -> that
			var addr int
			switch c.arg2 {
			case 0:
				addr = ptr["this"]
			case 1:
				addr = ptr["that"]
			default:
				return errors.New("wrong pointer")
			}
			b.popD()
			b.a(addr)
			b.c("M=D")
		default:
			return errors.New("wrong memory region")
		}
	case cmdPush:
		switch c.arg1 {
		case "constant":
			b.a(c.arg2)
			b.c("D=A")
		case "this", "that", "local", "argument":
			b.a(c.arg1)
			b.c("D=M") // diff
			b.a(c.arg2)
			b.c("A=A+D")
			b.c("D=M")
		case "temp":
			b.a(c.arg1)
			b.c("D=A") // diff
			b.a(c.arg2)
			b.c("A=A+D")
			b.c("D=M")
		case "pointer": // pointer 0 -> this, pointer 1 -> that
			var addr int
			switch c.arg2 {
			case 0:
				addr = ptr["this"]
			case 1:
				addr = ptr["that"]
			default:
				return errors.New("wrong pointer")
			}
			b.a(addr)
			b.c("D=M")
		case "static":
			b.c("@%s.%d", b.name, c.arg2)
			b.c("D=M")
		default:
			return errors.New("wrong memory region")
		}
		b.pushD()
	case cmdArithmetic: // after pop M stands for X, D stands for Y so  'x-y'  == 'm-y' i.e. subtract from later element on stack
		switch c.arg1 {
		case "add":
			b.popD()
			b.popM()
			b.c("D=M+D")
		case "and":
			b.popD()
			b.popM()
			b.c("D=D&M")
		case "or":
			b.popD()
			b.popM()
			b.c("D=D|M")
		case "not":
			b.popM()
			b.c("D=!M")
		case "neg":
			b.popM()
			b.c("D=-M")
		case "sub": // M - D
			b.popD()
			b.popM()
			b.c("D=M-D")
		case "eq", "gt", "lt":
			b.popD()
			b.popM()
			b.c("D=M-D")
			labelA, _ := b.l()
			b.a(labelA)
			switch c.arg1 {
			case "eq":
				b.c("D;JEQ")
			case "gt":
				b.c("D;JGT")
			case "lt":
				b.c("D;JLT")
			}
			b.c("D=0")
			labelB, _ := b.l()
			b.a(labelB)
			b.c("0;JMP")
			b.l(labelA)
			b.c("D=-1")
			b.l(labelB)
		default:
			return errors.New("not implemented")
		}
		b.pushD()
	default:
		return errors.New("not implemented")
	}
	return nil
}
