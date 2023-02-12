package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/a-h/rest/getcomments/parser"
)

var flagPkg = flag.String("pkg", "", "Name of the package to process.")
var flagOutput = flag.String("op", "", "Name of the file to write to.")

func main() {
	flag.Parse()
	if *flagPkg == "" {
		fmt.Println("missing package name")
		os.Exit(1)
	}
	if *flagOutput == "" {
		fmt.Println("missing output name")
		os.Exit(1)
	}
	m, err := parser.Get(*flagPkg)
	if err != nil {
		fmt.Printf("failed to get model: %v", err)
		os.Exit(1)
	}
	f, err := os.Create(*flagOutput)
	if err != nil {
		fmt.Printf("error creating output file %q: %v\n", *flagOutput, err)
		os.Exit(1)
	}
	fmt.Printf("snapshotting package %q\n", *flagPkg)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(m)
	if err != nil {
		fmt.Printf("error encoding: %v\n", err)
		os.Exit(1)
	}
}
