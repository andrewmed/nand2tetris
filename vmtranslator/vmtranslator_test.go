package vmtranslator

import (
	"bufio"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	in := `function Main.main 0
push constant 1
call Output.printInt 1
pop temp 0
push constant 0
return`

	out := `// function Main.main 0
(Main.main)
// push constant 1
@1
D=A
@0
A=M
M=D
@0
M=M+1
// call Output.printInt 1
@JUMP1
D=A
@0
A=M
M=D
@0
M=M+1
@1
D=M
@0
A=M
M=D
@0
M=M+1
@2
D=M
@0
A=M
M=D
@0
M=M+1
@3
D=M
@0
A=M
M=D
@0
M=M+1
@2
D=M
@0
A=M
M=D
@0
M=M+1
@0
D=M
@5
D=D-A
@1
D=D-A
@2
M=D
@0
D=M
@1
M=D
@Output.printInt
0;JMP
(JUMP1)
// pop temp 0
@5
D=A
@0
D=D+A
@13
M=D
@0
M=M-1
A=M
D=M
@13
A=M
M=D
// push constant 0
@0
D=A
@0
A=M
M=D
@0
M=M+1
// return
@1
D=M
@5
M=D
@5
D=D-A
@5
A=A+1
M=D
@0
M=M-1
A=M
D=M
@2
A=M
M=D
A=A+1
D=A
@0
M=D
@5
D=M
@1
A=D-A
D=M
@4
M=D
@5
D=M
@2
A=D-A
D=M
@3
M=D
@5
D=M
@3
A=D-A
D=M
@2
M=D
@5
D=M
@4
A=D-A
D=M
@1
M=D
@5
A=A+1
A=M
A=M
0;JMP
`

	b := NewAsmBuilder()
	reader := strings.NewReader(in)
	scanner := bufio.NewScanner(reader)
	TranslateFile(&b, scanner)

	if b.String() != out {
		t.Fatalf("want: %s, got: %s", out, b.String())
	}
}

