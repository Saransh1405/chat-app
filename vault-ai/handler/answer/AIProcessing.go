package answer

import (
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
	"strconv"
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
	historyQuery := `SELECT id, session_id, role, content, timestamp FROM chat_messages WHERE session_id = $1 ORDER BY timestamp ASC`
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

	limitStr := os.Getenv("SEARCH_RELEVANT_CHUNKS_LIMIT")
	limit := 10
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
			sourceName := chunk.Source
			if sourceName == "" {
				sourceName = "Unknown Document"
			}
			documentContext.WriteString(fmt.Sprintf("%d. [From: %s] %s\n\n", i+1, sourceName, chunk.Content))
		}
	}

	systemMessage := ollama.OllamaMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are a helpful AI assistant. 
Answer questions based on the user's conversation and their uploaded documents.
%s
Use the following context to answer the question:
If the answer is in the documents, cite them using [Source: Document Name] format at the end of the sentence or paragraph.
If the answer is not in the documents, state that you are using general knowledge.
Always be concise and helpful.`,
			documentContext.String()),
	}

	ChatCompletionMessage = append(ChatCompletionMessage, systemMessage)

	reqBody := ollama.OllamaRequest{
		Model:    "llama3.2",
		Messages: ChatCompletionMessage,
		Stream:   true,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error calling Ollama: %v", err)
		return
	}
	defer resp.Body.Close()

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk ollama.OllamaResponse
		if err := decoder.Decode(&chunk); err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error decoding: %v", err)
			break
		}

		fullResponse.WriteString(chunk.Message.Content)

		GlobalBroadcaster.BroadcastToSession(sessionID, chunk.Message.Content)

		if chunk.Done {
			break
		}
	}

	sessionManager.AddMessage(sessionID, "assistant", fullResponse.String())
	GlobalBroadcaster.BroadcastToSession(sessionID, "[DONE]")
}
