package proxy

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Usage represents the token usage information from API responses
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// parseUsageFromResponse extracts usage information from a non-streaming response body
func parseUsageFromResponse(body []byte) *Usage {
	if len(body) == 0 {
		return nil
	}

	var response struct {
		Usage *Usage `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil
	}

	return response.Usage
}

// parseUsageFromStreamChunks extracts usage from SSE stream data
// OpenAI-compatible APIs typically include usage in the last data chunk before [DONE]
func parseUsageFromStreamChunks(data []byte) *Usage {
	if len(data) == 0 {
		return nil
	}

	// Split by SSE data lines
	lines := bytes.Split(data, []byte("\n"))

	// Search from the end to find the last chunk with usage
	for i := len(lines) - 1; i >= 0; i-- {
		line := bytes.TrimSpace(lines[i])
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		jsonData := bytes.TrimPrefix(line, []byte("data:"))
		jsonData = bytes.TrimSpace(jsonData)

		// Skip [DONE] marker
		if bytes.Equal(jsonData, []byte("[DONE]")) {
			continue
		}

		var chunk struct {
			Usage *Usage `json:"usage"`
		}

		if err := json.Unmarshal(jsonData, &chunk); err != nil {
			continue
		}

		if chunk.Usage != nil && chunk.Usage.TotalTokens > 0 {
			return chunk.Usage
		}
	}

	return nil
}

// extractLastChunks keeps the last N bytes of stream data for usage extraction
// This is used to avoid storing the entire stream in memory
func extractLastChunks(buffer *bytes.Buffer, newData []byte, maxSize int) {
	buffer.Write(newData)

	// Keep only the last maxSize bytes
	if buffer.Len() > maxSize {
		data := buffer.Bytes()
		// Find a good break point (newline) to avoid cutting in the middle of a chunk
		start := buffer.Len() - maxSize
		for i := start; i < len(data); i++ {
			if data[i] == '\n' {
				start = i + 1
				break
			}
		}
		buffer.Reset()
		buffer.Write(data[start:])
	}
}

// isStreamingResponse checks if the response content-type indicates SSE streaming
func isStreamingResponse(contentType string) bool {
	return strings.Contains(contentType, "text/event-stream") ||
		strings.Contains(contentType, "application/x-ndjson")
}
