package answer

import (
	"bytes"
	"chat-app/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetDeepSeekEmbedding(question string) ([]float64, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}

	reqBody := models.EmbeddingRequest{
		Model: "deepseek-embedding",
		Input: question,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("body: ", string(body))
	fmt.Println("resp: ", resp)
	fmt.Println("resp.StatusCode: ", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", string(body))
		return nil, fmt.Errorf("failed to get embedding: %s", string(body))
	}

	var embeddingResp models.EmbeddingResponse
	err = json.Unmarshal(body, &embeddingResp)
	if err != nil {
		panic(err)
	}

	return embeddingResp.Data[0].Embedding, nil
}
