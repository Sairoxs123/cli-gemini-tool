package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai" // Import the Gemini client library
	"google.golang.org/api/option"
)

type Item struct {
	Name     string `json:"name"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	InitDone bool   `json:"init_done"`
}

func checkIfInArray(array []string, element string) bool {
	for _, val := range array {
		if val == element {
			return true
		}
	}
	return false
}

func readJSON() Item {
	filePath := "config.json"
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("The cli has not been initialized.")
	}
	var config Item
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v\n", err)
	}
	return config
}

func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

func initialize() {
	var config = readJSON()
	if config.InitDone {
		fmt.Println("You have already completed the initialization process.")
		return
	}
	fmt.Println("Initializing configuration...")

	name, err := readInput("What is your name (leave blank for anonymous): ")
	if err != nil {
		log.Printf("Warning: could not read name: %v\n", err)
		return
	}

	apiKey, err := readInput("Please enter your Gemini API key (compulsory): ")
	if err != nil {
		log.Fatalf("Fatal: could not read API key: %v\n", err)
		return
	}
	// Check length *after* trimming
	if len(apiKey) == 0 {
		fmt.Println("API key cannot be empty. It is required for the functioning of the CLI.")
		// Consider re-prompting or exiting cleanly
		return // Exit the initialize function
	}

	// --- Security Note ---
	// Storing API keys directly in a plain JSON is risky.
	// Consider environment variables or more secure storage.
	// Also, pasting keys into the terminal can expose them in shell history.

	model, err := readInput("Please enter the default Gemini model (e.g., gemini-1.5-flash-latest): ")
	if err != nil {
		log.Printf("Warning: could not read model name: %v\n", err)
		// Decide if you want to proceed with an empty model or exit
	}
	defaultModel := "gemini-2.0-flash-lite"
	if len(model) == 0 {
		fmt.Println("No default model entered. Setting gemini-2.0-flash-lite as the default model. Use blah blah --default <modelname> to change default model later.")
		model = defaultModel
	}

	// Use Go naming convention for struct fields if possible (ApiKey instead of API_key)
	// The json tag handles the mapping to the JSON field name.
	itemToWrite := Item{
		Name:     name,
		APIKey:   apiKey, // This field name in Go is conventional, json tag handles file format
		Model:    model,
		InitDone: true,
	}

	jsonData, err := json.MarshalIndent(itemToWrite, "", "  ")
	if err != nil {
		// Use log.Fatalf for errors that prevent continuation
		log.Fatalf("Error marshaling JSON: %v\n", err)
	}

	configFileName := "config.json"
	err = os.WriteFile(configFileName, jsonData, 0644) // 0644 are standard file permissions
	if err != nil {
		log.Fatalf("Error writing JSON file '%s': %v\n", configFileName, err)
	}

	fmt.Printf("Configuration saved successfully to %s\n", configFileName)
}

func sendMessage(ctx context.Context, cs *genai.ChatSession, userInput string) error {
	const bold = "\x1b[1m"
	fmt.Println("Gemini thinking...")
	// Send the user message and get the response
	resp, err := cs.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		// Return the error to be handled by the caller
		return fmt.Errorf("error sending message: %w", err)
	}

	// Print Gemini's response
	fmt.Print("Gemini: ")
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		// Iterate through the parts of the response content
		for _, part := range resp.Candidates[0].Content.Parts {
			// Print each part (usually just one for text)
			fmt.Println(strings.ReplaceAll(fmt.Sprintf("%v", part), "**", bold))
		}
		fmt.Println() // Add a newline after Gemini's full response
	} else {
		// Handle cases where the response might be empty or blocked
		fmt.Println("Received no response content.")
	}
	return nil // Indicate success
}

// The main function is where the program execution begins.
func main() {
	var config Item = readJSON()
	ctx := context.Background()
	apiKey := config.APIKey
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}
	model := client.GenerativeModel(config.Model)
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}
	cs := model.StartChat()

	// Check if any arguments were provided (os.Args[0] is the program name)
	if len(os.Args) > 1 {
		// Simple command checking
		commands := []string{"--init", "init", "-f", "-i", "--set-api-key", "--set-model", "--output=json", "summarize", "code-explainer"}
		if !checkIfInArray(commands, os.Args[1]) {
			sendMessage(ctx, cs, os.Args[1])
		} else {
			if strings.ToLower(os.Args[1]) == "--init" || strings.ToLower(os.Args[1]) == "init" {
				initialize()
			} else {
				fmt.Printf("Unknown command: %s\n", os.Args[1])
				fmt.Println("Available commands: init") // List available commands
			}
		}
	} else {
		// No command was provided
		fmt.Println("Please provide a command (e.g., 'init').")
		// You could print usage instructions here
	}
	defer client.Close()
}
