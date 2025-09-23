package proxy

import (
	"bytes"
	"io"
	"net/http"

	"gpt-load/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ResponseData 包含响应数据和token使用量
type ResponseData struct {
	Body       []byte
	TokenUsage *TokenUsage
}

func (ps *ProxyServer) handleStreamingResponse(c *gin.Context, resp *http.Response, group *models.Group) *ResponseData {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		logrus.Error("Streaming unsupported by the writer, falling back to normal response")
		return ps.handleNormalResponse(c, resp, group)
	}

	var accumulatedData bytes.Buffer
	parser := GetTokenParser(group.ChannelType)

	buf := make([]byte, 4*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			accumulatedData.Write(chunk)

			if _, writeErr := c.Writer.Write(chunk); writeErr != nil {
				logUpstreamError("writing stream to client", writeErr)
				return &ResponseData{Body: accumulatedData.Bytes()}
			}
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logUpstreamError("reading from upstream", err)
			return &ResponseData{Body: accumulatedData.Bytes()}
		}
	}

	responseBody := accumulatedData.Bytes()

	// 尝试从流式数据中解析 token 使用量
	var tokenUsage *TokenUsage
	if usage, err := parser.ParseStreamingTokenUsage(responseBody); err == nil {
		tokenUsage = usage
	}

	return &ResponseData{
		Body:       responseBody,
		TokenUsage: tokenUsage,
	}
}

func (ps *ProxyServer) handleNormalResponse(c *gin.Context, resp *http.Response, group *models.Group) *ResponseData {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logUpstreamError("reading response body", err)
		return &ResponseData{}
	}

	parser := GetTokenParser(group.ChannelType)
	var tokenUsage *TokenUsage

	if usage, parseErr := parser.ParseTokenUsage(body); parseErr == nil {
		tokenUsage = usage
	} else {
		logrus.Debugf("Failed to parse token usage: %v", parseErr)
	}

	if _, writeErr := c.Writer.Write(body); writeErr != nil {
		logUpstreamError("writing response body", writeErr)
	}

	return &ResponseData{
		Body:       body,
		TokenUsage: tokenUsage,
	}
}
