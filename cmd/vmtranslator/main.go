package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"git.andmed.org/nand2tetris/vmtranslator"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("Usage: vmtranslator /path/to/fileORdir")
	}
	path := os.Args[1]

	stat, e := os.Stat(path)
	if e != nil {
		log.Fatal(e)
	}

	b := vmtranslator.NewAsmBuilder()
	vmtranslator.Bootstrap(&b)

	var files int
	lines := 0
	if stat.IsDir() {
		filenames, _ := filepath.Glob(path + "/*.vm")
		for _, filename := range filenames {
			files++
			file, _ := os.Open(filename)
			scanner := bufio.NewScanner(file)
			lines += vmtranslator.TranslateFile(&b, scanner)
		}
	} else {
		files = 1
		file, _ := os.Open(path)
		scanner := bufio.NewScanner(file)
		lines = vmtranslator.TranslateFile(&b, scanner)
	}

	fmt.Print(b.String())
	log.Printf("Total %d lines in %d VM files processed.\n", lines, files)
	os.Exit(vmtranslator.ErrFound)
}
