# Solutions for https://www.nand2tetris.org/ ![Build status](https://travis-ci.org/andrewmed/nand2tetris.svg?branch=master) ![Go report card](https://goreportcard.com/badge/github.com/andrewmed/nand2tetris)


## CPU.hdl and Computer.hdl
Main hardware parts of a computer, ran HACK machine code
- Implement register-based "chain of computation"

## HackAssembler (java)
Translates assembly (ASM) to machine code (HACK), double pass

## vmtranslator (go)
Translates (VM) code to assembly (ASM), single pass
- Implements stack based computations
- Implements function call conventions
- Does "linking" 

## compiler (go)
Compiles high level (JACK) code into intermediate representation (VM), tree parsing

All parts can be tested in Hardware emulator and CPU emulator by the link above

Copyright: MIT Andmed, 2019 (c)