// author andmed
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const comments = true

type cmdType int

const (
	C_Arithmetic cmdType = iota
	C_Push
	C_Pop
	C_Label
	C_Goto
	C_If
	C_Function
	C_Return
	C_Call
)

const (
	SP       = "SP"
	local    = "local"
	argument = "argument"
	this     = "this"
	that     = "that"
	temp     = "temp"
)

var (
	vmName   string
	errFound int
)

// pointers to mem indexes
var P = map[string]int{
	SP:       0, // stack pointer
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

type asmBuilder struct {
	jumpIndex int // jump label index
	strings.Builder
}

func newAsmBuilder() asmBuilder {
	return asmBuilder{0, strings.Builder{}}
}

func main() {

	if len(os.Args) != 2 {
		log.Fatal("Usage: vmtranslator /path/to/fileORdir")
	}
	path := os.Args[1]

	stat, e := os.Stat(path)
	if e != nil {
		log.Fatal(e)
	}

	b := newAsmBuilder()
	bootstrap(&b)

	var files int
	lines := 0
	if stat.IsDir() {
		filenames, _ := filepath.Glob(path + "/*.vm")
		for _, filename := range filenames {
			files++
			file, _ := os.Open(filename)
			lines += translateFile(&b, file)
		}
	} else {
		files = 1
		file, _ := os.Open(path)
		lines = translateFile(&b, file)
	}

	fmt.Print(b.String())
	log.Printf("Total %d lines in %d VM files processed.\n", lines, files)
	os.Exit(errFound)
}

func bootstrap(b *asmBuilder) {
	log.Println("Compiling bootstrap code")
	// set SP
	if comments {
		b.c("// bootstrap section")
	}
	b.a(256)
	b.c("D=A")
	b.a(SP)
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

func translateFile(b *asmBuilder, file *os.File) int {
	log.Println("Compiling " + file.Name())
	scanner := bufio.NewScanner(file)
	//compile := regexp.MustCompile(`.*/(\w+)\.vm`)
	//vmName = compile.FindStringSubmatch(filename)[1]
	//compile.MatchString(filename)
	//log.Println("compiling " + vmName)
	line := 0
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
			errFound++
			continue
		}
		if e := generate(b, c); e != nil {
			log.Printf("Line %d %s: '%s'\n", line, e, s)
			errFound++
			continue
		}
	}
	log.Printf("%d lines processed.\n", line)
	return line
}

func parse(s string) (cmdStruct, error) {
	tokens := strings.Split(s, " ")

	switch tokens[0] {
	case "function":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: C_Function, arg1: tokens[1], arg2: arg2}, nil
	case "call":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: C_Call, arg1: tokens[1], arg2: arg2}, nil
	case "goto":
		return cmdStruct{cmd: C_Goto, arg1: tokens[1]}, nil
	case "if-goto":
		return cmdStruct{cmd: C_If, arg1: tokens[1]}, nil
	case "return":
		return cmdStruct{cmd: C_Return}, nil
	case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
		return cmdStruct{cmd: C_Arithmetic, arg1: tokens[0]}, nil
	case "push":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: C_Push, arg1: tokens[1], arg2: arg2}, nil
	case "pop":
		arg2, _ := strconv.Atoi(tokens[2])
		return cmdStruct{cmd: C_Pop, arg1: tokens[1], arg2: arg2}, nil
	case "label":
		return cmdStruct{cmd: C_Label, arg1: tokens[1]}, nil
	}

	return cmdStruct{}, errors.New("parsing")
}

// A command
// if arg is int adds "@arg"
// if str arg is in  P{} add it as pointer val
// otherwise just add as it is (var ref)
func (b *asmBuilder) a(x interface{}) (int, error) {
	switch v := x.(type) {
	case int:
		return b.WriteString(fmt.Sprintf("@%d\n", v))
	case string:
		i, ok := P[v]
		if !ok {
			return b.WriteString(fmt.Sprintf("@%s\n", v))
		} else {
			return b.WriteString(fmt.Sprintf("@%d\n", i))
		}
	default:
		panic("wrong b.a() usage")
	}
}

// B command
// with one arg just adds
// with more formats in Sprintf
func (b *asmBuilder) c(s ...interface{}) (int, error) {
	format := s[0].(string) + "\n"
	if len(s) == 1 {
		return b.WriteString(format)
	} else {
		return b.WriteString(fmt.Sprintf(format, s[1:]...))
	}
}

// if no arg, returns next jump label
// otherwise, outputs asm label
func (b *asmBuilder) l(s ...string) (string, error) {
	if len(s) == 0 {
		b.jumpIndex++
		return "JUMP" + strconv.Itoa(b.jumpIndex), nil
	}
	_, e := b.WriteString("(" + s[0] + ")\n")
	return "", e
}

// pop from stack to M register
func (b *asmBuilder) popM() (int, error) {
	return b.WriteString("@0\nM=M-1\nA=M\n")
}

// push D register on stack
func (b *asmBuilder) pushD() (int, error) {
	return b.WriteString("@0\nA=M\nM=D\n@0\nM=M+1\n")
}

// pop from stack to D register
func (b *asmBuilder) popD() (int, error) {
	return b.WriteString("@0\nM=M-1\nA=M\nD=M\n")
}

func generate(b *asmBuilder, c cmdStruct) error {
	switch c.cmd {
	case C_Call:
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
		// ARG = SP - (n of args) - 5
		b.a(SP)
		b.c("D=M")
		b.a(5)
		b.c("D=D-A")
		b.a(c.arg2)
		b.c("D=D-A")
		b.a(argument)
		b.c("M=D")
		// LCL= SP
		b.a(SP)
		b.c("D=M")
		b.a(local)
		b.c("M=D")
		// goto f
		b.a(c.arg1)
		b.c("0;JMP")
		// ret label
		b.l(label)

	case C_Return:
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
		// SP = *ARG + 1
		b.c("A=A+1")
		b.c("D=A")
		b.a(SP)
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
	case C_If:
		b.popD()
		b.a(c.arg1)
		b.c("D;JNE")
	case C_Goto:
		b.a(c.arg1)
		b.c("0;JMP")
	case C_Label:
		b.c("(%s)", c.arg1)
	case C_Function:
		b.c("(%s)", c.arg1)
		for ; c.arg2 > 0; c.arg2-- { // preallocate locals
			b.pushD() // no need to clear, this should be job of c higher level language
		}
	case C_Pop: // pop from stack
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
			b.c("@%s.%d", vmName, c.arg2)
			b.c("M=D")
		case "pointer": // pointer 0 -> this, pointer 1 -> that
			var addr int
			switch c.arg2 {
			case 0:
				addr = P["this"]
			case 1:
				addr = P["that"]
			default:
				return errors.New("wrong pointer")
			}
			b.popD()
			b.a(addr)
			b.c("M=D")
		default:
			return errors.New("wrong memory region")
		}
	case C_Push:
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
				addr = P["this"]
			case 1:
				addr = P["that"]
			default:
				return errors.New("wrong pointer")
			}
			b.a(addr)
			b.c("D=M")
		case "static":
			b.c("@%s.%d", vmName, c.arg2)
			b.c("D=M")
		default:
			return errors.New("wrong memory region")
		}
		b.pushD()
	case C_Arithmetic: // after pop M stands for X, D stands for Y so  'x-y'  == 'm-y' i.e. subtract from later element on stack
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
