package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"tale/blocks"
	"tale/parser"
)

func findNestedTalePaths(absDirPath string) []string {
	var talePaths []string

	entries, err := os.ReadDir(absDirPath)
	if err != nil {
		return talePaths
	}

	for _, entry := range entries {
		absPath := path.Join(absDirPath, entry.Name())

		if entry.IsDir() {
			talePaths = append(talePaths, findNestedTalePaths(absPath)...)
		} else if path.Ext(entry.Name()) == ".tale" {
			talePaths = append(talePaths, absPath)
		}
	}

	return talePaths
}

func streamTaleFile(absTalePath string, blockChan chan blocks.Block) {
	p := parser.New(absTalePath)
	block := p.Next()

	for block.Type != blocks.END_OF_BLOCKS {
		blockChan <- block
		block = p.Next()
	}

	blockChan <- block
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var dirPaths, talePaths []string

	for _, arg := range os.Args[1:] {
		absPath := arg
		if !path.IsAbs(absPath) {
			absPath = path.Join(pwd, absPath)
		}

		switch path.Ext(absPath) {
		case ".tale":
			talePaths = append(talePaths, absPath)
		case "":
			dirPaths = append(dirPaths, absPath)
		default:
			log.Fatalf("Error: Path must be a directory or .tale file %q\n", arg)
		}
	}

	if len(dirPaths) == 0 && len(talePaths) == 0 {
		dirPaths = append(dirPaths, pwd)
	}

	for _, dirPath := range dirPaths {
		talePaths = append(talePaths, findNestedTalePaths(dirPath)...)
	}

	if len(talePaths) == 0 {
		log.Fatal("Error: No .tale files found!")
	}

	blockChan := make(chan blocks.Block)

	for _, talePath := range talePaths {
		go streamTaleFile(talePath, blockChan)
	}

	var streamedBlocks []blocks.Block

	for doneCount := 0; doneCount < len(talePaths); {
		block :=  <-blockChan

		if block.Type == blocks.END_OF_BLOCKS {
			doneCount += 1
		} else {
			streamedBlocks = append(streamedBlocks, block)
		}
	}

	fmt.Printf("Blocks:\n%q\n", streamedBlocks)
}
