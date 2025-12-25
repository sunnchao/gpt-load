package proxy

import (
	"bytes"
	"io"
	"net/http"

	"gpt-load/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// maxUsageBufferSize is the maximum size of the buffer used to extract usage from streaming responses
const maxUsageBufferSize = 16 * 1024

func (ps *ProxyServer) handleStreamingResponse(c *gin.Context, resp *http.Response) *Usage {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		logrus.Error("Streaming unsupported by the writer, falling back to normal response")
		return ps.handleNormalResponse(c, resp)
	}

	// Buffer to store the last chunks for usage extraction
	var usageBuffer bytes.Buffer

	buf := make([]byte, 4*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				logUpstreamError("writing stream to client", writeErr)
				return nil
			}
			flusher.Flush()

			// Accumulate chunks for usage extraction
			extractLastChunks(&usageBuffer, buf[:n], maxUsageBufferSize)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logUpstreamError("reading from upstream", err)
			return nil
		}
	}

	// Parse usage from the accumulated stream data
	return parseUsageFromStreamChunks(usageBuffer.Bytes())
}

func (ps *ProxyServer) handleNormalResponse(c *gin.Context, resp *http.Response) *Usage {
	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logUpstreamError("reading response body", err)
		return nil
	}

	// Decompress if needed
	contentEncoding := resp.Header.Get("Content-Encoding")
	decompressed, err := utils.DecompressResponse(contentEncoding, bodyBytes)
	if err != nil {
		logrus.WithError(err).Warn("Failed to decompress response, using raw body")
		decompressed = bodyBytes
	}

	// Parse usage from the response
	usage := parseUsageFromResponse(decompressed)

	// Write the original body to the client
	if _, writeErr := c.Writer.Write(bodyBytes); writeErr != nil {
		logUpstreamError("writing response body", writeErr)
	}

	return usage
}
