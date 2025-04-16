package main

/*
This example shows how to use a chat template to generate a prompt for a model and tokenize it.

The chat template is extracted from the tokenizer_config.json file.

required environment variables:
- CHAT_TEMPLATE_PATH: path to tokenizer_config.json containing the chat template
- TOKENIZER_PATH: path to tokenizer.json

*/

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/tech-arch1tect/tokenizers-cpp-go/manual"
)

// chat template config is tokenizer_config.json, but all we are using is the chat template
type ChatTemplateConfig struct {
	ChatTemplate string `json:"chat_template"`
}

// chat message is a single message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	// get the tokenizer config
	// this is a tokenizer_config.json, but all we are using is the chat template
	chatTplPath := os.Getenv("CHAT_TEMPLATE_PATH")
	if chatTplPath == "" {
		log.Fatal("Please set CHAT_TEMPLATE_PATH")
	}
	chatBytes, err := os.ReadFile(chatTplPath)
	if err != nil {
		log.Fatalf("Reading chat template config: %v", err)
	}
	var chatCfg ChatTemplateConfig
	if err := json.Unmarshal(chatBytes, &chatCfg); err != nil {
		log.Fatalf("Parsing chat template JSON: %v", err)
	}
	// end of tokenizer config

	// define the conversation
	chat := []ChatMessage{
		{Role: "system", Content: "You are a knowledgeable assistant."},
		{Role: "user", Content: "Tell me about Go's concurrency model."},
		{Role: "assistant", Content: "Go uses goroutines and channels to manage concurrency."},
		{Role: "user", Content: "How do they differ from OS threads?"},
	}
	// end of conversation

	// render the chat template
	var msgs []map[string]interface{}
	for _, m := range chat {
		msgs = append(msgs, map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	ctx := exec.NewContext(map[string]interface{}{
		"messages": msgs,
	})

	tpl, err := gonja.FromString(chatCfg.ChatTemplate)
	if err != nil {
		log.Fatalf("Compiling Jinja template with gonja/v2: %v", err)
	}
	var renderedPrompt strings.Builder
	if err := tpl.Execute(&renderedPrompt, ctx); err != nil {
		log.Fatalf("Rendering chat prompt: %v", err)
	}

	fmt.Println("Rendered Chat Prompt:")
	fmt.Println(renderedPrompt.String())
	// end of rendering the chat template

	// get the tokenizer
	tokPath := os.Getenv("TOKENIZER_PATH")
	if tokPath == "" {
		log.Fatal("Please set TOKENIZER_PATH")
	}
	tokBytes, err := os.ReadFile(tokPath)
	if err != nil {
		log.Fatalf("Reading tokenizer config: %v", err)
	}
	// end of getting the tokenizer

	// generate tokens from the rendered prompt
	tokenizer, err := manual.NewFromJSON(string(tokBytes))
	if err != nil {
		log.Fatalf("Initialising tokenizer: %v", err)
	}
	defer tokenizer.Free()

	tokenIDs, err := tokenizer.Encode(renderedPrompt.String(), true)
	if err != nil {
		log.Fatalf("Tokenising rendered prompt: %v", err)
	}
	fmt.Println("\nToken IDs:")
	fmt.Println(tokenIDs)
	// end of generating tokens from the rendered prompt
}
