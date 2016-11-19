package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bruston/roost/parser"
	"github.com/bruston/roost/runtime"
)

func main() {
	var input io.ReadCloser
	if len(os.Args) < 2 {
		input = os.Stdin
	} else {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatalf("unable to open source: %s", err)
		}
		input = file
	}
	defer input.Close()
	env := runtime.New(1024)
	var p *parser.Parser
	if len(os.Args) >= 2 {
		p = parser.NewParser(input)
		ast, err := p.Parse()
		if err != nil {
			log.Fatal(err)
		}
		runtime.Eval(env, ast)
		return
	}

	// else start the REPL
	fmt.Print("repl> ")
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		src := scanner.Bytes()
		p = parser.NewParser(bytes.NewReader(src))
		ast, err := p.Parse()
		if err != nil {
			log.Print(err)
		}
		runtime.Eval(env, ast)
		fmt.Printf("\nrepl> ")
	}
	if scanner.Err() != nil {
		fmt.Fprintf(os.Stderr, "%s", scanner.Err())
	}
}
