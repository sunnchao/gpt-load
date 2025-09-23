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
	CacheCreationTokens    int `json:"cache_creation_tokens,omitempty"`
	CacheReadTokens        int `json:"cache_read_tokens,omitempty"`
	Ephemeral5mTokens      int `json:"ephemeral_5m_tokens,omitempty"`
	Ephemeral1hTokens      int `json:"ephemeral_1h_tokens,omitempty"`
	ReasoningTokens        int `json:"reasoning_tokens,omitempty"`
	AudioTokens            int `json:"audio_tokens,omitempty"`
	ImageTokens            int `json:"image_tokens,omitempty"`
	ServiceTier            string `json:"service_tier,omitempty"`
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
	// Claude streaming format includes usage in message_start and message_delta events
	lines := bytes.Split(data, []byte("\n"))

	var messageStartUsage *TokenUsage
	var messageDeltaUsage *TokenUsage

	for i := 0; i < len(lines); i++ {
		line := bytes.TrimSpace(lines[i])
		if bytes.HasPrefix(line, []byte("data: ")) {
			jsonData := bytes.TrimPrefix(line, []byte("data: "))

			// Parse message_start event for initial usage info
			var messageStartEvent struct {
				Type    string `json:"type"`
				Message *struct {
					ID    string `json:"id"`
					Model string `json:"model"`
					Usage struct {
						InputTokens               int `json:"input_tokens"`
						OutputTokens              int `json:"output_tokens"`
						CacheCreationInputTokens  int `json:"cache_creation_input_tokens,omitempty"`
						CacheReadInputTokens      int `json:"cache_read_input_tokens,omitempty"`
						CacheCreation            *struct {
							Ephemeral5mInputTokens  int `json:"ephemeral_5m_input_tokens,omitempty"`
							Ephemeral1hInputTokens  int `json:"ephemeral_1h_input_tokens,omitempty"`
						} `json:"cache_creation,omitempty"`
						ServiceTier              string `json:"service_tier,omitempty"`
					} `json:"usage"`
				} `json:"message,omitempty"`
			}

			if err := json.Unmarshal(jsonData, &messageStartEvent); err == nil {
				if messageStartEvent.Type == "message_start" && messageStartEvent.Message != nil {
					usage := messageStartEvent.Message.Usage
					totalTokens := usage.InputTokens + usage.OutputTokens
					cachedTokens := usage.CacheCreationInputTokens + usage.CacheReadInputTokens

					// Add ephemeral cache tokens if present
					ephemeral5m := 0
					ephemeral1h := 0
					if usage.CacheCreation != nil {
						ephemeral5m = usage.CacheCreation.Ephemeral5mInputTokens
						ephemeral1h = usage.CacheCreation.Ephemeral1hInputTokens
						cachedTokens += ephemeral5m + ephemeral1h
					}

					messageStartUsage = &TokenUsage{
						PromptTokens:        usage.InputTokens,
						CompletionTokens:    usage.OutputTokens,
						TotalTokens:         totalTokens,
						CachedPromptTokens:  cachedTokens,
						CacheCreationTokens: usage.CacheCreationInputTokens,
						CacheReadTokens:     usage.CacheReadInputTokens,
						Ephemeral5mTokens:   ephemeral5m,
						Ephemeral1hTokens:   ephemeral1h,
						ServiceTier:         usage.ServiceTier,
					}
				}
			}

			// Parse message_delta event for final usage info
			var messageDeltaEvent struct {
				Type  string `json:"type"`
				Delta *struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta,omitempty"`
				Usage *struct {
					InputTokens               int `json:"input_tokens"`
					OutputTokens              int `json:"output_tokens"`
					CacheCreationInputTokens  int `json:"cache_creation_input_tokens,omitempty"`
					CacheReadInputTokens      int `json:"cache_read_input_tokens,omitempty"`
					CacheCreation            *struct {
						Ephemeral5mInputTokens  int `json:"ephemeral_5m_input_tokens,omitempty"`
						Ephemeral1hInputTokens  int `json:"ephemeral_1h_input_tokens,omitempty"`
					} `json:"cache_creation,omitempty"`
				} `json:"usage,omitempty"`
			}

			if err := json.Unmarshal(jsonData, &messageDeltaEvent); err == nil {
				if messageDeltaEvent.Type == "message_delta" && messageDeltaEvent.Usage != nil {
					usage := messageDeltaEvent.Usage
					totalTokens := usage.InputTokens + usage.OutputTokens
					cachedTokens := usage.CacheCreationInputTokens + usage.CacheReadInputTokens

					// Add ephemeral cache tokens if present
					ephemeral5m := 0
					ephemeral1h := 0
					if usage.CacheCreation != nil {
						ephemeral5m = usage.CacheCreation.Ephemeral5mInputTokens
						ephemeral1h = usage.CacheCreation.Ephemeral1hInputTokens
						cachedTokens += ephemeral5m + ephemeral1h
					}

					messageDeltaUsage = &TokenUsage{
						PromptTokens:        usage.InputTokens,
						CompletionTokens:    usage.OutputTokens,
						TotalTokens:         totalTokens,
						CachedPromptTokens:  cachedTokens,
						CacheCreationTokens: usage.CacheCreationInputTokens,
						CacheReadTokens:     usage.CacheReadInputTokens,
						Ephemeral5mTokens:   ephemeral5m,
						Ephemeral1hTokens:   ephemeral1h,
					}
				}
			}

			// Also handle legacy message_stop format
			var messageStopEvent struct {
				Type    string `json:"type"`
				Message *struct {
					Usage struct {
						InputTokens               int `json:"input_tokens"`
						OutputTokens              int `json:"output_tokens"`
						CacheCreationInputTokens  int `json:"cache_creation_input_tokens,omitempty"`
						CacheReadInputTokens      int `json:"cache_read_input_tokens,omitempty"`
					} `json:"usage"`
				} `json:"message,omitempty"`
			}

			if err := json.Unmarshal(jsonData, &messageStopEvent); err == nil {
				if messageStopEvent.Type == "message_stop" && messageStopEvent.Message != nil {
					usage := messageStopEvent.Message.Usage
					totalTokens := usage.InputTokens + usage.OutputTokens
					cachedTokens := usage.CacheCreationInputTokens + usage.CacheReadInputTokens

					return &TokenUsage{
						PromptTokens:       usage.InputTokens,
						CompletionTokens:   usage.OutputTokens,
						TotalTokens:        totalTokens,
						CachedPromptTokens: cachedTokens,
					}, nil
				}
			}
		}
	}

	// Return the most complete usage info available
	// Prefer message_delta over message_start as it contains final token counts
	if messageDeltaUsage != nil {
		return messageDeltaUsage, nil
	}
	if messageStartUsage != nil {
		return messageStartUsage, nil
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