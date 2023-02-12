package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/a-h/rest/getcomments/parser"
)

func main() {
	m, err := parser.Get("github.com/a-h/rest/getcomments/parser/example")
	if err != nil {
		log.Fatalf("failed to parse: %v", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(m)
	if err != nil {
		fmt.Printf("error encoding: %v\n", err)
		os.Exit(1)
	}
}
