package answer

import (
	"bytes"
	"chat-app/internal/models"
	"chat-app/vault-ai/library/ollama"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"chat-app/internal/database"

	"github.com/google/uuid"
)

func AIProcessing(ctx context.Context, sessionID uuid.UUID, question string, userID uuid.UUID, db *database.DB) {
	// 1. Get session with ALL previous messages
	sessionManager := NewSessionManager(db)
	_, err := sessionManager.GetSession(userID, sessionID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	// 2. Load ALL messages (history)
	var messages []models.ChatMessage
	query := `SELECT * FROM chat_messages WHERE session_id = $1 ORDER BY timestamp ASC`
	err = db.QueryRow(query, sessionID).Scan(&messages)
	if err != nil {
		log.Printf("Error loading messages: %v", err)
		return
	}

	// 3. Build context for llm
	var ChatCompletionMessage []ollama.OllamaMessage

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
	// get the limit from the environment variable
	limitStr := os.Getenv("SEARCH_RELEVANT_CHUNKS_LIMIT")
	limit := 10 // default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		} else {
			log.Printf("Error parsing SEARCH_RELEVANT_CHUNKS_LIMIT: %v, using default value %d", err, limit)
		}
	}

	relevantChunks, err := SearchRelevantChunks(ctx, question, userID, limit, db)
	if err != nil {
		log.Printf("Error searching chunks: %v", err)
	}

	var documentContext strings.Builder
	if len(relevantChunks) > 0 {
		documentContext.WriteString("Relevant information from your documents:\n\n")
		for i, chunk := range relevantChunks {
			documentContext.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, chunk.Content))
		}
	}

	systemMessage := ollama.OllamaMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are a helpful AI assistant. 
Answer questions based on the user's conversation and their uploaded documents.
%s
Use the following context to answer the question:
If the answer is in the documents, cite them. If not, use your general knowledge.`,
			documentContext.String()),
	}

	ChatCompletionMessage = append(ChatCompletionMessage, systemMessage)

	// Create request
	reqBody := ollama.OllamaRequest{
		Model:    "llama3.2",
		Messages: ChatCompletionMessage,
		Stream:   true,
	}

	jsonData, _ := json.Marshal(reqBody)

	// Create request
	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error calling Ollama: %v", err)
		return
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	// Read streaming response
	for {
		var chunk ollama.OllamaResponse
		if err := decoder.Decode(&chunk); err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error decoding: %v", err)
			break
		}

		fullResponse.WriteString(chunk.Message.Content)

		// Broadcast to clients
		GlobalBroadcaster.BroadcastToSession(sessionID, chunk.Message.Content)

		if chunk.Done {
			break
		}
	}

	// Save response
	sessionManager.AddMessage(sessionID, "assistant", fullResponse.String())
	GlobalBroadcaster.BroadcastToSession(sessionID, "[DONE]")
}
