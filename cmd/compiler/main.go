package main

import (
	"git.andmed.org/nand2tetris/compiler"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: compiler /path/to/fileORdir")
	}
	path := os.Args[1]

	stat, e := os.Stat(path)
	if e != nil {
		log.Fatal(e)
	}

	var files int
	if stat.IsDir() {
		filenames, _ := filepath.Glob(path + "/*.jack")
		for _, filename := range filenames {
			compiler.CompilePath(filename)
			files++
		}
	} else {
		files = 1
		compiler.CompilePath(path)
	}

	log.Printf("Total %d files processed.\n", files)

}
