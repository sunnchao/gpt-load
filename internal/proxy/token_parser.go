// Package proxy provides token usage parsing for different AI providers
package proxy

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

// TokenUsage represents comprehensive token usage information
type TokenUsage struct {
	PromptTokens           int `json:"prompt_tokens"`
	CompletionTokens       int `json:"completion_tokens"`
	TotalTokens            int `json:"total_tokens"`
	CachedPromptTokens     int `json:"cached_prompt_tokens,omitempty"`
	CachedCompletionTokens int `json:"cached_completion_tokens,omitempty"`
	ReasoningTokens        int `json:"reasoning_tokens,omitempty"`
	AudioTokens            int `json:"audio_tokens,omitempty"`
	ImageTokens            int `json:"image_tokens,omitempty"`
}

// ResponseParser interface for different AI providers
type ResponseParser interface {
	ParseTokenUsage(body []byte) (*TokenUsage, error)
	ParseStreamingTokenUsage(data []byte) (*TokenUsage, error)
}

// OpenAIParser handles OpenAI and compatible API responses
type OpenAIParser struct{}

func (p *OpenAIParser) ParseTokenUsage(body []byte) (*TokenUsage, error) {
	var response struct {
		Usage struct {
			PromptTokens                   int `json:"prompt_tokens"`
			CompletionTokens               int `json:"completion_tokens"`
			TotalTokens                    int `json:"total_tokens"`
			PromptTokensDetails            *struct {
				CachedTokens int `json:"cached_tokens,omitempty"`
				AudioTokens  int `json:"audio_tokens,omitempty"`
			} `json:"prompt_tokens_details,omitempty"`
			CompletionTokensDetails        *struct {
				ReasoningTokens int `json:"reasoning_tokens,omitempty"`
				AudioTokens     int `json:"audio_tokens,omitempty"`
			} `json:"completion_tokens_details,omitempty"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	usage := &TokenUsage{
		PromptTokens:     response.Usage.PromptTokens,
		CompletionTokens: response.Usage.CompletionTokens,
		TotalTokens:      response.Usage.TotalTokens,
	}

	// Parse cached tokens
	if response.Usage.PromptTokensDetails != nil {
		usage.CachedPromptTokens = response.Usage.PromptTokensDetails.CachedTokens
		usage.AudioTokens += response.Usage.PromptTokensDetails.AudioTokens
	}

	// Parse reasoning tokens
	if response.Usage.CompletionTokensDetails != nil {
		usage.ReasoningTokens = response.Usage.CompletionTokensDetails.ReasoningTokens
		usage.AudioTokens += response.Usage.CompletionTokensDetails.AudioTokens
	}

	return usage, nil
}

func (p *OpenAIParser) ParseStreamingTokenUsage(data []byte) (*TokenUsage, error) {
	// OpenAI streaming format: data: {"usage": {...}}
	lines := bytes.Split(data, []byte("\n"))

	// Look for usage information in the last chunks
	for i := len(lines) - 1; i >= 0; i-- {
		line := bytes.TrimSpace(lines[i])
		if bytes.HasPrefix(line, []byte("data: ")) {
			jsonData := bytes.TrimPrefix(line, []byte("data: "))
			if bytes.Equal(jsonData, []byte("[DONE]")) {
				continue
			}

			var chunk struct {
				Usage *struct {
					PromptTokens                   int `json:"prompt_tokens"`
					CompletionTokens               int `json:"completion_tokens"`
					TotalTokens                    int `json:"total_tokens"`
					PromptTokensDetails            *struct {
						CachedTokens int `json:"cached_tokens,omitempty"`
						AudioTokens  int `json:"audio_tokens,omitempty"`
					} `json:"prompt_tokens_details,omitempty"`
					CompletionTokensDetails        *struct {
						ReasoningTokens int `json:"reasoning_tokens,omitempty"`
						AudioTokens     int `json:"audio_tokens,omitempty"`
					} `json:"completion_tokens_details,omitempty"`
				} `json:"usage,omitempty"`
			}

			if err := json.Unmarshal(jsonData, &chunk); err == nil && chunk.Usage != nil {
				usage := &TokenUsage{
					PromptTokens:     chunk.Usage.PromptTokens,
					CompletionTokens: chunk.Usage.CompletionTokens,
					TotalTokens:      chunk.Usage.TotalTokens,
				}

				if chunk.Usage.PromptTokensDetails != nil {
					usage.CachedPromptTokens = chunk.Usage.PromptTokensDetails.CachedTokens
					usage.AudioTokens += chunk.Usage.PromptTokensDetails.AudioTokens
				}

				if chunk.Usage.CompletionTokensDetails != nil {
					usage.ReasoningTokens = chunk.Usage.CompletionTokensDetails.ReasoningTokens
					usage.AudioTokens += chunk.Usage.CompletionTokensDetails.AudioTokens
				}

				return usage, nil
			}
		}
	}

	return nil, nil
}

// GeminiParser handles Google Gemini API responses
type GeminiParser struct{}

func (p *GeminiParser) ParseTokenUsage(body []byte) (*TokenUsage, error) {
	var response struct {
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
			CachedContentTokenCount int `json:"cachedContentTokenCount,omitempty"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &TokenUsage{
		PromptTokens:           response.UsageMetadata.PromptTokenCount,
		CompletionTokens:       response.UsageMetadata.CandidatesTokenCount,
		TotalTokens:            response.UsageMetadata.TotalTokenCount,
		CachedPromptTokens:     response.UsageMetadata.CachedContentTokenCount,
	}, nil
}

func (p *GeminiParser) ParseStreamingTokenUsage(data []byte) (*TokenUsage, error) {
	// Gemini doesn't typically provide usage in streaming, try to find in final chunk
	return p.ParseTokenUsage(data)
}

// ClaudeParser handles Anthropic Claude API responses
type ClaudeParser struct{}

func (p *ClaudeParser) ParseTokenUsage(body []byte) (*TokenUsage, error) {
	var response struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	totalTokens := response.Usage.InputTokens + response.Usage.OutputTokens
	cachedTokens := response.Usage.CacheCreationInputTokens + response.Usage.CacheReadInputTokens

	return &TokenUsage{
		PromptTokens:       response.Usage.InputTokens,
		CompletionTokens:   response.Usage.OutputTokens,
		TotalTokens:        totalTokens,
		CachedPromptTokens: cachedTokens,
	}, nil
}

func (p *ClaudeParser) ParseStreamingTokenUsage(data []byte) (*TokenUsage, error) {
	// Claude streaming format includes usage in message_stop event
	lines := bytes.Split(data, []byte("\n"))

	for i := len(lines) - 1; i >= 0; i-- {
		line := bytes.TrimSpace(lines[i])
		if bytes.HasPrefix(line, []byte("data: ")) {
			jsonData := bytes.TrimPrefix(line, []byte("data: "))

			var event struct {
				Type    string `json:"type"`
				Message *struct {
					Usage struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
						CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
						CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
					} `json:"usage"`
				} `json:"message,omitempty"`
			}

			if err := json.Unmarshal(jsonData, &event); err == nil {
				if event.Type == "message_stop" && event.Message != nil {
					totalTokens := event.Message.Usage.InputTokens + event.Message.Usage.OutputTokens
					cachedTokens := event.Message.Usage.CacheCreationInputTokens + event.Message.Usage.CacheReadInputTokens

					return &TokenUsage{
						PromptTokens:       event.Message.Usage.InputTokens,
						CompletionTokens:   event.Message.Usage.OutputTokens,
						TotalTokens:        totalTokens,
						CachedPromptTokens: cachedTokens,
					}, nil
				}
			}
		}
	}

	return nil, nil
}

// GetTokenParser returns the appropriate parser for the given channel type
func GetTokenParser(channelType string) ResponseParser {
	switch strings.ToLower(channelType) {
	case "openai", "azure-openai", "azure_openai":
		return &OpenAIParser{}
	case "gemini", "google", "google-gemini":
		return &GeminiParser{}
	case "anthropic", "claude":
		return &ClaudeParser{}
	default:
		// Default to OpenAI parser for unknown types
		return &OpenAIParser{}
	}
}

// Helper function to extract usage from generic streaming response
func ExtractUsageFromStreamData(data []byte, channelType string) (*TokenUsage, error) {
	parser := GetTokenParser(channelType)
	return parser.ParseStreamingTokenUsage(data)
}

// Helper function to search for usage patterns in streaming data
func findUsageInStreamingData(data []byte) *TokenUsage {
	// Generic regex patterns for common usage formats
	patterns := []string{
		`"usage":\s*{[^}]*"total_tokens":\s*(\d+)[^}]*"prompt_tokens":\s*(\d+)[^}]*"completion_tokens":\s*(\d+)`,
		`"prompt_tokens":\s*(\d+)[^}]*"completion_tokens":\s*(\d+)[^}]*"total_tokens":\s*(\d+)`,
		`"input_tokens":\s*(\d+)[^}]*"output_tokens":\s*(\d+)`,
	}

	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			if matches := re.FindSubmatch(data); len(matches) >= 3 {
				// Try to parse the numeric values
				// This is a simplified approach - in practice, you'd want more robust parsing
				break
			}
		}
	}

	return nil
}