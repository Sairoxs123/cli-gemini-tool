package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath" // Added for potential path cleaning
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai" // Import the Gemini client library
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

const version string = "v1.0.0"

// Struct to hold user configuration
type Item struct {
	Name     string `json:"name"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	InitDone bool   `json:"init_done"`
}

// Map common language identifiers to file extensions
var languageExtensions = map[string]string{
	"python":     "py",
	"javascript": "js",
	"java":       "java",
	"c":          "c",
	"cpp":        "cpp",
	"csharp":     "cs",
	"go":         "go",
	"ruby":       "rb",
	"php":        "php",
	"swift":      "swift",
	"kotlin":     "kt",
	"typescript": "ts",
	"html":       "html",
	"css":        "css",
	"json":       "json",
	"xml":        "xml",
	"yaml":       "yaml",
	"sql":        "sql",
	"bash":       "sh",
	"rust":       "rs",
	// Add more as needed
}

// Helper to check if an element exists in a string slice
func checkIfInArray(array []string, element string) bool {
	for _, val := range array {
		if val == element {
			return true
		}
	}
	return false
}

// Reads configuration from config.json
func readJSON() Item {
	filePath := "config.json"
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist, return an empty Item to allow initialization
		if os.IsNotExist(err) {
			fmt.Println("config.json not found. Run 'init' command.")
			return Item{} // Return empty struct, InitDone will be false
		}
		log.Fatalf("Error reading config file %s: %v\n", filePath, err)
	}
	var config Item
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON from %s: %v\n", filePath, err)
	}
	return config
}

// Reads a line of input from the user
func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

// Initializes the configuration file
func initialize() {
	config := readJSON() // Read existing or get empty struct
	if config.InitDone {
		fmt.Println("You have already completed the initialization process.")
		// Optionally ask if they want to re-initialize
		return
	}
	fmt.Println("Initializing configuration...")

	name, err := readInput("What is your name (leave blank for anonymous): ")
	if err != nil {
		log.Printf("Warning: could not read name: %v\n", err)
		// Decide if this is fatal or recoverable
	}

	apiKey, err := readInput("Please enter your Gemini API key (required): ")
	if err != nil || len(apiKey) == 0 {
		log.Fatal("API key is required and could not be read.")
	}

	model, err := readInput("Enter default Gemini model (e.g., gemini-2.0-flash-lite) [leave blank for default]: ")
	if err != nil {
		log.Printf("Warning: could not read model name: %v\n", err)
	}
	if len(model) == 0 {
		model = "gemini-2.0-flash-lite" // Updated default model
		fmt.Printf("Using default model: %s\n", model)
	}

	itemToWrite := Item{
		Name:     name,
		APIKey:   apiKey,
		Model:    model,
		InitDone: true,
	}

	jsonData, err := json.MarshalIndent(itemToWrite, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v\n", err)
	}

	configFileName := "config.json"
	err = os.WriteFile(configFileName, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file '%s': %v\n", configFileName, err)
	}

	fmt.Printf("Configuration saved successfully to %s\n", configFileName)
}

// Struct to hold extracted code block information
type CodeBlock struct {
	Language string
	Content  string
	// Store the start and end position of the block in the original text
	StartIndex int
	EndIndex   int
}

// Regular expression to find code blocks and capture their content and language.
// It now also captures the full match including the backticks.
var codeBlockRegex = regexp.MustCompile("(?s)(\x60\x60\x60(?:\\w*)\n?(.*?)\n?\x60\x60\x60)")

// Regular expression to find filename directives *after* code blocks.
// Case-insensitive, handles optional bold markers and backticks around filename.
var filenameDirectiveRegex = regexp.MustCompile("(?i)(?:\\*\\*)?Filename:(?:\\*\\*)?\\s*`?([^`\n]+)`?")

// Extracts code blocks (```...```) from a string.
// Returns a slice of CodeBlock structs including their positions.
func extractCodeBlocksWithPositions(text string) []CodeBlock {
	matches := codeBlockRegex.FindAllStringSubmatchIndex(text, -1) // Find all matches and their indices

	var blocks []CodeBlock

	if matches == nil {
		return blocks // Return empty slice if no matches found
	}

	// Submatch indices for the compiled regex:
	// idx[0], idx[1]: Full match (```lang\ncontent```)
	// idx[2], idx[3]: Group 1 (full match again, needed for overall position)
	// idx[4], idx[5]: Group 2 (content inside the block)

	for _, idx := range matches {
		if len(idx) >= 6 { // Need indices for full match and content group
			fullBlockText := text[idx[0]:idx[1]]
			content := text[idx[4]:idx[5]]

			// Extract language from the start of the full block text
			firstLineEnd := strings.Index(fullBlockText, "\n")
			language := ""
			if firstLineEnd > 3 { // Check if there's anything after ```
				// Extract potential language identifier between ``` and the first newline
				langPart := strings.TrimSpace(fullBlockText[3:firstLineEnd])
				// Basic check if it looks like a valid language identifier (alphanumeric, no spaces/backticks)
				isValidLang := true
				if len(langPart) == 0 {
					isValidLang = false
				}
				for _, r := range langPart {
					if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
						isValidLang = false
						break
					}
				}
				if isValidLang {
					language = langPart
				}
			} // Note: No explicit handling for ```lang content``` on a single line, assumes ```lang\ncontent``` format primarily

			blocks = append(blocks, CodeBlock{
				Language:   language,
				Content:    strings.TrimSpace(content),
				StartIndex: idx[0], // Start index of the full ```...``` block
				EndIndex:   idx[1], // End index of the full ```...``` block
			})
		}
	}

	return blocks
}

// Helper function to get text from a genai.Part
func getTextFromPart(part genai.Part) (string, bool) {
	if textPart, ok := part.(genai.Text); ok {
		return string(textPart), true
	}
	return "", false
}

// Processes a response to extract code blocks and save them to files,
// attempting to find filenames from directives following the blocks.
func processResponseAndSaveFiles(resp *genai.GenerateContentResponse) {
	if resp == nil || resp.Candidates == nil || len(resp.Candidates) == 0 {
		log.Println("Received an empty response")
		return
	}
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		log.Println("Response candidate has no content parts")
		return
	}

	// Concatenate text from all parts (usually just one)
	var fullResponseTextBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		if text, ok := getTextFromPart(part); ok {
			fullResponseTextBuilder.WriteString(text)
		}
	}
	fullResponseText := fullResponseTextBuilder.String()

	if fullResponseText == "" {
		log.Println("Response contains no text parts.")
		return
	}

	// --- Extract Code Blocks ---
	codeBlocks := extractCodeBlocksWithPositions(fullResponseText)

	if len(codeBlocks) == 0 {
		// log.Println("No code blocks found in the response.") // Optional: Only log if expecting code
		return // Nothing to save
	}

	log.Println("--- Processing Code Blocks ---")
	// Removed unused variable: lastIndexProcessed

	for i, block := range codeBlocks {
		fmt.Printf("Found Block %d: Language='%s'\n", i+1, block.Language)

		// --- Determine Filename ---
		filename := ""
		// Search for the directive in the text *after* the current block ends
		// and *before* the next block starts (or until the end of the text).
		searchStart := block.EndIndex
		searchEnd := len(fullResponseText)
		if i+1 < len(codeBlocks) { // If there's a next block, limit search area
			searchEnd = codeBlocks[i+1].StartIndex
		}

		// Ensure searchStart is not past searchEnd (can happen with adjacent blocks)
		if !(searchStart >= searchEnd) {
			searchText := fullResponseText[searchStart:searchEnd]
			directiveMatch := filenameDirectiveRegex.FindStringSubmatch(searchText)

			if len(directiveMatch) > 1 {
				// Group 1 contains the captured filename
				potentialFilename := strings.TrimSpace(directiveMatch[1])
				if potentialFilename != "" {
					// Basic sanitization (replace common problematic chars) - more robust needed for production
					sanitizedFilename := strings.ReplaceAll(potentialFilename, "/", "_")
					sanitizedFilename = strings.ReplaceAll(sanitizedFilename, "\\", "_")
					// Use filepath.Clean for potentially better path handling, though might be overkill here
					filename = filepath.Clean(sanitizedFilename)
					fmt.Printf("  Found directive, using filename: %s\n", filename)
				}
			}
		}

		// --- Fallback to UUID if no filename found ---
		if filename == "" {
			ext := languageExtensions[strings.ToLower(block.Language)]
			if ext == "" {
				ext = "txt" // Default to .txt if language unknown or not mapped
			}
			filename = uuid.NewString() + "." + ext
			fmt.Printf("  No filename directive found, using generated name: %s\n", filename)
		}

		if strings.Contains(block.Content, "(rest of the") {
			continue
		}

		err := os.WriteFile(filename, []byte(block.Content), 0644) // Use block.Content
		if err != nil {
			fmt.Printf("  Error writing file '%s': %v\n", filename, err)
		} else {
			fmt.Printf("  Successfully saved content to %s\n", filename)
		}
		// No need to update lastIndexProcessed here
	}
	log.Println("--- Finished Processing Code Blocks ---")
}

// Sends a message (text or text+file) to the chat session and prints the response.
func sendMessage(ctx context.Context, cs *genai.ChatSession, client *genai.Client, userInput string, filePath string) error {
	var resp *genai.GenerateContentResponse
	var err error
	parts := []genai.Part{genai.Text(userInput)} // Start with the text prompt

	// --- Handle File Upload if filePath is provided ---
	if filePath != "" {
		// Ensure file exists before attempting upload
		if _, errStat := os.Stat(filePath); os.IsNotExist(errStat) {
			return fmt.Errorf("file not found: %s", filePath)
		}

		fmt.Printf("Uploading file: %s...\n", filePath)
		// Use the client passed to the function for uploading
		uploadedFile, errUpload := client.UploadFileFromPath(ctx, filePath, nil)
		if errUpload != nil {
			// Attempt to use UploadFileFromPath as a fallback or alternative if needed
			// uploadedFile, errUpload = client.UploadFileFromPath(ctx, filePath, nil)
			// if errUpload != nil {
			return fmt.Errorf("error uploading file '%s': %w", filePath, errUpload)
			// }
		}
		fmt.Printf("File uploaded successfully! URI: %s\n", uploadedFile.URI)

		// Prepend file data to the parts slice
		filePart := genai.FileData{
			MIMEType: uploadedFile.MIMEType, // Use MIME type from upload result
			URI:      uploadedFile.URI,
		}
		parts = append([]genai.Part{filePart}, parts...) // Add file part at the beginning
	}

	// --- Send Message ---
	fmt.Println("You:", userInput) // Echo user input
	fmt.Println("Gemini thinking...")
	resp, err = cs.SendMessage(ctx, parts...)
	if err != nil {
		// Return the error to be handled by the caller
		return fmt.Errorf("error sending message: %w", err)
	}

	// --- Print Gemini's Response ---
	fmt.Print("Gemini: ")
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			fmt.Print(strings.ReplaceAll(fmt.Sprintf("%v", part), "**", "\x1b[1m")) // Use Print to avoid extra newlines between parts if multiple exist
		}
		fmt.Println()
		processResponseAndSaveFiles(resp)

	} else {
		fmt.Println("Received no response content or response was blocked.")
		if resp.PromptFeedback != nil {
			fmt.Printf("Prompt Feedback Block Reason: %s\n", resp.PromptFeedback.BlockReason.String())
			for _, rating := range resp.PromptFeedback.SafetyRatings {
				fmt.Printf("  Safety Rating - Category: %s, Probability: %s\n", rating.Category.String(), rating.Probability.String())
			}
		}
		if len(resp.Candidates) > 0 && resp.Candidates[0].FinishReason != genai.FinishReasonStop {
			fmt.Printf("Candidate Finish Reason: %s\n", resp.Candidates[0].FinishReason.String())
		}
	}

	return nil
}

func setModel(model string, models *genai.ModelInfoIterator) {
	temp_models := []string{}
	for {
		modelInfo, err := models.Next()
		if err != nil {
			break
		}
		temp_models = append(temp_models, modelInfo.Name)
	}

	if !checkIfInArray(temp_models, model) {
		log.Fatalf("Invalid model name: %s\n", model)
		fmt.Printf("Available Models are: ")
		for _, val := range temp_models {
			fmt.Printf("%s\n", val)
		}
		return
	}
	config := readJSON()
	config.Model = model
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v\n", err)
	}
	configFileName := "config.json"
	err = os.WriteFile(configFileName, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file '%s': %v\n", configFileName, err)
	}

	fmt.Printf("Default model successfully set to %s\n", model)
}

func setAPIKey(apiKey string) {
	config := readJSON()
	config.APIKey = apiKey
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v\n", err)
	}
	configFileName := "config.json"
	err = os.WriteFile(configFileName, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON file '%s': %v\n", configFileName, err)
	}

	fmt.Printf("API Key updated successfully.\n")
}

func main() {
	config := readJSON()
	if !config.InitDone && (len(os.Args) <= 1 || (os.Args[1] != "init" && os.Args[1] != "--init")) {
		log.Fatal("CLI not initialized. Please run the 'init' command first.")
	}

	ctx := context.Background()

	// Handle init command separately as it doesn't need the client yet
	if len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "--init") {
		initialize()
		return // Exit after initialization
	}

	// --- Initialize Gemini Client ---
	apiKey := config.APIKey
	if apiKey == "" {
		log.Fatal("API key is missing in config.json. Please run 'init' again.")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating Gemini client: %v\n", err)
	}
	defer client.Close() // Ensure client is closed when main exits

	modelName := config.Model
	if modelName == "" {
		modelName = "gemini-2.0-flash-lite" // Fallback default
		log.Printf("Model name missing in config, using default: %s\n", modelName)
	}
	model := client.GenerativeModel(modelName)
	cs := model.StartChat() // Start a chat session

	// --- Argument Parsing and Command Handling ---
	if len(os.Args) > 1 {
		command := os.Args[1]
		filePath := ""
		prompt := ""

		// Very basic argument parsing - consider using the 'flag' package for robust CLI apps
		if command == "-f" || command == "--file" {
			if len(os.Args) < 3 {
				log.Fatal("File path missing after -f/--file flag.")
			}
			filePath = os.Args[2]
			if len(os.Args) > 3 {
				prompt = strings.Join(os.Args[3:], " ") // Join remaining args as prompt
				// Handle specific sub-commands for files
				if prompt == "summarize" {
					prompt = "Provide a concise summary of the content of this file."
				} else if prompt == "code-explainer" {
					prompt = "Explain the code in this file step-by-step, focusing on its purpose and key logic."
				}
			} else {
				// If only -f <file> is given, ask for a default prompt
				prompt = "Describe the contents of this file."
				fmt.Printf("No specific prompt provided for file '%s'. Using default prompt: '%s'\n", filePath, prompt)
			}
		} else if command == "--set-api-key" {
			setAPIKey(os.Args[2])
			return
		} else if command == "--set-model" {
			setModel(strings.ToLower(os.Args[2]), client.ListModels(ctx))
			return
		} else if command == "-chat" {
			// --- Interactive Mode (Example) ---
			fmt.Println("Entering interactive chat mode (type 'exit' or 'quit' to end).")
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("You: ")
				userInput, _ := reader.ReadString('\n')
				userInput = strings.TrimSpace(userInput)

				if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" {
					fmt.Println("Exiting chat.")
					break
				}

				if userInput == "" {
					continue
				}

				err := sendMessage(ctx, cs, client, userInput, "") // No file in interactive mode for simplicity
				if err != nil {
					log.Printf("Error during chat: %v", err)
					// Decide if error is fatal or if chat can continue
				}
			}
		} else if command == "--help" {
			fmt.Println("Available commands:")
			fmt.Println("  init: Initialize the CLI tool (required before first use).")
			fmt.Println("  -f <file_path> [prompt]: Send a file to Gemini with an optional prompt.")
			fmt.Println("  --file <file_path> [prompt]: Same as -f.")
			fmt.Println("  --set-api-key <api_key>: Set or update the Gemini API key.")
			fmt.Println("  --set-model <model_name>: Set the default Gemini model.")
			fmt.Println("  -chat: Enter interactive chat mode.")
			fmt.Println("  [prompt]: Send a text prompt to Gemini.")
			fmt.Println("  -f <file_path> summarize: Summarize the content of the file.")
			fmt.Println("  -f <file_path> code-explainer: Explain the code in the file.")
			fmt.Println("  --help: Gives list of commands available.")
			fmt.Println("  --version: Prints the CLI version.")
			return
		} else if command == "--version" {
			fmt.Printf("Gemini CLI Version: %s\n", version)
		} else {
			// Assume the first argument onwards is the prompt
			prompt = strings.Join(os.Args[1:], " ")
		}

		// Send the message (potentially with file)
		if prompt != "" {
			err := sendMessage(ctx, cs, client, prompt, filePath) // Pass client for uploads
			if err != nil {
				log.Fatalf("Error: %v", err) // Log fatal errors from sending
			}
		} else {
			log.Println("No prompt provided.")
		}

	} else {
		fmt.Println("Available commands:")
		fmt.Println("  init: Initialize the CLI tool (required before first use).")
		fmt.Println("  -f <file_path> [prompt]: Send a file to Gemini with an optional prompt.")
		fmt.Println("  --file <file_path> [prompt]: Same as -f.")
		fmt.Println("  --set-api-key <api_key>: Set or update the Gemini API key.")
		fmt.Println("  --set-model <model_name>: Set the default Gemini model.")
		fmt.Println("  -chat: Enter interactive chat mode.")
		fmt.Println("  [prompt]: Send a text prompt to Gemini.")
		fmt.Println("  -f <file_path> summarize: Summarize the content of the file.")
		fmt.Println("  -f <file_path> code-explainer: Explain the code in the file.")
	}
}
