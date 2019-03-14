package main

import (
	"fmt"
	"git.andmed.org/nand2tetris/vmtranslator"
	"log"
	"os"
	"path/filepath"
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

	b := vmtranslator.VMTranslator{}
	vmtranslator.Bootstrap(&b)

	var files, lines, exitCode int
	var ok bool
	if stat.IsDir() {
		filenames, _ := filepath.Glob(path + "/*.vm")
		for _, filename := range filenames {
			files++
			line, ok := vmtranslator.TranslateFile(&b, filename)
			lines += line
			if !ok {
				exitCode = 1
			}

		}
	} else {
		files = 1
		lines, ok = vmtranslator.TranslateFile(&b, path)
		if !ok {
			exitCode = 1
		}
	}

	fmt.Print(b.String())
	log.Printf("Total %d lines in %d VM files processed.\n", lines, files)
	os.Exit(exitCode)
}
