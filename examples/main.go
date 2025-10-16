package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/matthiase/alpaca/bindings"
)

func main() {
	modelPath := flag.String("model", "", "Path to GGUF model")
	flag.Parse()

	if *modelPath == "" {
		log.Fatal("Please provide -model flag")
	}

	// Initialize backend
	bindings.Init()
	defer bindings.Free()

	// Load model
	fmt.Println("Loading model...")
	model, err := bindings.LoadModel(*modelPath)
	if err != nil {
		log.Fatal(err)
	}
	defer model.Free()

	// Print basic info
	fmt.Printf("✓ Model loaded successfully!\n")
	fmt.Printf("  Vocabulary size: %d\n", model.VocabSize())
	fmt.Printf("  Context size: %d\n", model.ContextSize())
}
