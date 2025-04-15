package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tech-arch1tect/tokenizers-cpp-go/manual"
)

func main() {
	tokenIDs, err := getTokens()
	if err != nil {
		log.Fatalf("failed to get tokens: %v", err)
	}

	fmt.Println("Token IDs:")
	for _, id := range tokenIDs {
		fmt.Println(id)
	}
}

func getTokens() ([]int32, error) {
	tokenizerPath := os.Getenv("TOKENIZER_PATH")
	if tokenizerPath == "" {
		return nil, fmt.Errorf("TOKENIZER_PATH environment variable is not set")
	}

	jsonBytes, err := os.ReadFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tokenizer configuration file: %v", err)
	}
	jsonConfig := string(jsonBytes)

	tokenizer, err := manual.NewFromJSON(jsonConfig)
	if err != nil {
		return nil, fmt.Errorf("initialising tokenizer: %v", err)
	}
	defer tokenizer.Free()

	inputText := "Hello, how are you?"
	tokenIDs, err := tokenizer.Encode(inputText, true)
	if err != nil {
		return nil, fmt.Errorf("encoding input: %v", err)
	}

	return tokenIDs, nil
}
