# Solutions for https://www.nand2tetris.org/


## CPU.hdl and Computer.hdl
Main hardware parts of a computer
- Implement register-based "chain of computation"


## HackAssembler.java
Translates assembley to machine code, mostly just line by line (except handling labels, where double pass is needed)

## vmtranslator.go
Translates VM code to assembler
- Implements stack based computations (translates to a "flat" assembly code)
- Implements function call conventions

All parts can be tested in Hardware emulator and CPU emulator by the link above



Copyright: MIT Andmed, 2019 (c)


