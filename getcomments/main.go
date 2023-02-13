package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/a-h/rest/getcomments/parser"
)

var flagPackage = flag.String("package", "", "The package to retrieve comments from, e.g. github.com/a-h/rest/getcomments/example")

func main() {
	flag.Parse()
	if *flagPackage == "" {
		flag.Usage()
		os.Exit(0)
	}
	m, err := parser.Get(*flagPackage)
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
