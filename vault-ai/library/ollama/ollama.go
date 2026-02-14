package ollama

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type OllamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEuE/QSFiBF/i9lSPFIQp5LU4nvnfQyQ00dSSatO+d+S
