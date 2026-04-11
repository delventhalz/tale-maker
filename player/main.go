package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"tale/lexer"
	"tale/tokens"
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

	for _, talePath := range talePaths {
		taleBytes, err := os.ReadFile(talePath)
		if err != nil {
			log.Fatal(err)
		}

		lex := lexer.New(string(taleBytes))
		for tok := lex.Next(); tok.Type != tokens.EOF; tok = lex.Next() {
			fmt.Println(tok)
		}

		fmt.Println("\n")
	}
}
