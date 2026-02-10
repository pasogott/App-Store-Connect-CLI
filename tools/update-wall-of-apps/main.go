package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/wallgen"
)

func main() {
	repoRoot, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	result, err := wallgen.Generate(repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Updated %s\n", result.GeneratedPath)
	fmt.Printf("Synced snippet markers in %s\n", result.ReadmePath)
}
