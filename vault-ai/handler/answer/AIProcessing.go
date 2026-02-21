package answer

import (
	"bufio"
	"bytes"
	"chat-app/internal/database"
	"chat-app/internal/models"
	"chat-app/vault-ai/library/ollama"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func AIProcessing(ctx context.Context, sessionID uuid.UUID, question string, userID uuid.UUID, db *database.DB) {
	sessionManager := NewSessionManager(db)
	_, err := sessionManager.GetSession(userID, sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	var messages []models.ChatMessage
	historyQuery := `SELECT id, session_id, role, content, timestamp FROM ai_chat_messages WHERE session_id = $1 ORDER BY timestamp ASC`
	rows, err := db.Query(historyQuery, sessionID)
	if err != nil {
		log.Printf("Error loading messages: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Timestamp); err != nil {
			log.Printf("Error scanning history message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// System message should be first
	systemMessage := ollama.OllamaMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are a helpful AI assistant. 
Answer questions based on the user's conversation and their uploaded documents.
%s
Use the following context to answer the question:
If the answer is in the documents, cite them using [Source: Document Name] format at the end of the sentence or paragraph.
If the answer is not in the documents, state that you are using general knowledge.
Always be concise and helpful.`,
			"",
		),
	}

	var ChatCompletionMessage []ollama.OllamaMessage
	ChatCompletionMessage = append(ChatCompletionMessage, systemMessage)

	// Add conversation history
	for _, msg := range messages {
		if msg.Role == "user" {
			ChatCompletionMessage = append(ChatCompletionMessage, ollama.OllamaMessage{
				Role:    "user",
				Content: msg.Content,
			})
		} else {
			ChatCompletionMessage = append(ChatCompletionMessage, ollama.OllamaMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		}
	}

	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "llama3:latest"
	}

	reqBody := ollama.OllamaRequest{
		Model:    modelName,
		Messages: ChatCompletionMessage,
		Stream:   true,
	}

	jsonData, _ := json.Marshal(reqBody)
	log.Printf("Sending request to Ollama: %s", string(jsonData))

	GlobalBroadcaster.BroadcastToSession(sessionID, "[TYPING]")

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollamaEndpoint := fmt.Sprintf("%s/api/chat", ollamaURL)
	log.Printf("Ollama endpoint: %s", ollamaEndpoint)

	// Create HTTP request for streaming
	req, err := http.NewRequest("POST", ollamaEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Use HTTP client - no timeout for streaming
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling Ollama: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("resp.StatusCode: %v\n", resp.StatusCode)
	log.Printf("Response headers - Content-Type: %s, Transfer-Encoding: %s",
		resp.Header.Get("Content-Type"), resp.Header.Get("Transfer-Encoding"))

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
		return
	}

	var fullResponse strings.Builder

	reader := bufio.NewReaderSize(resp.Body, 64*1024)
	lineCount := 0

	log.Printf("Starting to read streaming response...")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Reached EOF after %d lines", lineCount)
				break
			}
			log.Printf("Error reading line: %v", err)
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			log.Printf("Skipping empty line")
			continue
		}

		lineCount++
		log.Printf("Line %d from Ollama (length: %d): %s", lineCount, len(line), line)

		var chunk ollama.OllamaResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			log.Printf("Error unmarshaling chunk: %v, line: %s", err, line)
			continue
		}

		log.Printf("Parsed chunk %d: Content='%s', Done=%v", lineCount, chunk.Message.Content, chunk.Done)
		fmt.Printf("chunk: %+v\n", chunk)

		if chunk.Message.Content != "" {
			fullResponse.WriteString(chunk.Message.Content)
			fmt.Println("chunk.Message.Content: ", chunk.Message.Content)
			GlobalBroadcaster.BroadcastToSession(sessionID, chunk.Message.Content)
		}

		if chunk.Done {
			log.Printf("Received done flag at line %d, breaking", lineCount)
			break
		}
	}

	log.Printf("Finished reading response, total lines: %d, full response length: %d", lineCount, fullResponse.Len())
	sessionManager.AddMessage(sessionID, "assistant", fullResponse.String())
	GlobalBroadcaster.BroadcastToSession(sessionID, "[DONE]")
}
