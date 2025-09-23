import type { RequestLog } from "@/types/models";

/**
 * Claude API streaming response event types
 */
export interface ClaudeStreamEvent {
  type: string;
  [key: string]: any;
}

export interface ClaudeMessageStart extends ClaudeStreamEvent {
  type: "message_start";
  message: {
    id: string;
    type: "message";
    role: "assistant";
    model: string;
    content: any[];
    stop_reason: null;
    stop_sequence: null;
    usage: ClaudeTokenUsage;
    service_tier: string;
  };
}

export interface ClaudeMessageDelta extends ClaudeStreamEvent {
  type: "message_delta";
  delta: {
    stop_reason: string | null;
    stop_sequence: null;
  };
  usage: ClaudeTokenUsage;
}

export interface ClaudeTokenUsage {
  input_tokens: number;
  cache_creation_input_tokens?: number;
  cache_read_input_tokens?: number;
  cache_creation?: {
    ephemeral_5m_input_tokens?: number;
    ephemeral_1h_input_tokens?: number;
  };
  output_tokens: number;
  service_tier?: string;
}

export interface TokenConsumptionRecord {
  timestamp: string;
  requestId: string;
  model: string;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  cacheCreationTokens?: number;
  cacheReadTokens?: number;
  cachedPromptTokens?: number;
  ephemeral5mTokens?: number;
  ephemeral1hTokens?: number;
  reasoningTokens?: number;
  audioTokens?: number;
  imageTokens?: number;
  serviceTier?: string;
  rawResponse?: string;
}

/**
 * Claude Token Tracker Service
 * 解析Claude API流式响应并提取token使用量信息
 */
export class ClaudeTokenTracker {
  private tokenRecords: TokenConsumptionRecord[] = [];
  private storageKey = "claude_token_consumption";

  /**
   * 解析Claude API流式响应文本
   */
  parseStreamingResponse(responseText: string): TokenConsumptionRecord | null {
    try {
      const events = this.parseStreamEvents(responseText);
      if (events.length === 0) return null;

      const messageStart = events.find(e => e.type === "message_start") as ClaudeMessageStart;
      const messageDelta = events.find(e => e.type === "message_delta") as ClaudeMessageDelta;

      if (!messageStart) return null;

      // Use final usage from message_delta if available, otherwise use message_start
      const finalUsage = messageDelta?.usage || messageStart.message.usage;
      const startUsage = messageStart.message.usage;

      // Calculate cached tokens including cache_creation
      let cachedTokens = (finalUsage.cache_read_input_tokens || 0) +
                        (finalUsage.cache_creation_input_tokens || 0);

      // Extract ephemeral cache tokens
      const ephemeral5m = finalUsage.cache_creation?.ephemeral_5m_input_tokens || 0;
      const ephemeral1h = finalUsage.cache_creation?.ephemeral_1h_input_tokens || 0;

      // Add ephemeral cache tokens to total cached tokens
      cachedTokens += ephemeral5m + ephemeral1h;

      return {
        timestamp: new Date().toISOString(),
        requestId: messageStart.message.id,
        model: messageStart.message.model,
        inputTokens: finalUsage.input_tokens || 0,
        outputTokens: finalUsage.output_tokens || 0,
        totalTokens: (finalUsage.input_tokens || 0) + (finalUsage.output_tokens || 0),
        cacheCreationTokens: finalUsage.cache_creation_input_tokens,
        cacheReadTokens: finalUsage.cache_read_input_tokens,
        cachedPromptTokens: cachedTokens,
        ephemeral5mTokens: ephemeral5m,
        ephemeral1hTokens: ephemeral1h,
        serviceTier: messageStart.message.service_tier,
        rawResponse: responseText
      };
    } catch (error) {
      console.error("Failed to parse Claude streaming response:", error);
      return null;
    }
  }

  /**
   * 解析流式事件数据
   */
  private parseStreamEvents(responseText: string): ClaudeStreamEvent[] {
    const events: ClaudeStreamEvent[] = [];
    const lines = responseText.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();

      // Handle SSE format: "event: message_start" followed by "data: {...}"
      if (line.startsWith('event:')) {
        const eventType = line.replace('event:', '').trim();

        // 查找对应的data行
        for (let j = i + 1; j < lines.length; j++) {
          const dataLine = lines[j].trim();
          if (dataLine.startsWith('data:')) {
            try {
              const dataContent = dataLine.replace('data:', '').trim();
              if (dataContent &&
                  dataContent !== '{"type": "ping"}' &&
                  !dataContent.includes('"type":"ping"')) {
                const eventData = JSON.parse(dataContent);
                eventData.type = eventType;
                events.push(eventData);
              }
            } catch (e) {
              console.warn("Failed to parse event data:", dataLine);
            }
            break;
          } else if (dataLine.startsWith('event:')) {
            // Found next event, stop looking for data
            break;
          }
        }
      }
      // Also handle direct "data: {...}" lines without preceding "event:" lines
      else if (line.startsWith('data:')) {
        try {
          const dataContent = line.replace('data:', '').trim();
          if (dataContent &&
              dataContent !== '{"type": "ping"}' &&
              !dataContent.includes('"type":"ping"')) {
            const eventData = JSON.parse(dataContent);
            // Only add if it has a type field
            if (eventData.type) {
              events.push(eventData);
            }
          }
        } catch (e) {
          console.warn("Failed to parse data line:", line);
        }
      }
    }

    return events;
  }

  /**
   * 存储token消费记录
   */
  storeTokenRecord(record: TokenConsumptionRecord): void {
    this.tokenRecords.push(record);
    this.saveToStorage();
  }

  /**
   * 批量处理Claude API响应并存储
   */
  processAndStore(responseText: string): TokenConsumptionRecord | null {
    const record = this.parseStreamingResponse(responseText);
    if (record) {
      this.storeTokenRecord(record);
    }
    return record;
  }

  /**
   * 获取所有token消费记录
   */
  getTokenRecords(): TokenConsumptionRecord[] {
    return [...this.tokenRecords];
  }

  /**
   * 获取统计信息
   */
  getUsageStats(): {
    totalTokens: number;
    totalInputTokens: number;
    totalOutputTokens: number;
    totalCachedTokens: number;
    recordCount: number;
    modelBreakdown: Record<string, { count: number; tokens: number }>;
  } {
    const records = this.getTokenRecords();

    const stats = {
      totalTokens: 0,
      totalInputTokens: 0,
      totalOutputTokens: 0,
      totalCachedTokens: 0,
      recordCount: records.length,
      modelBreakdown: {} as Record<string, { count: number; tokens: number }>
    };

    records.forEach(record => {
      stats.totalTokens += record.totalTokens;
      stats.totalInputTokens += record.inputTokens;
      stats.totalOutputTokens += record.outputTokens;
      stats.totalCachedTokens += record.cachedPromptTokens || 0;

      if (!stats.modelBreakdown[record.model]) {
        stats.modelBreakdown[record.model] = { count: 0, tokens: 0 };
      }
      stats.modelBreakdown[record.model].count++;
      stats.modelBreakdown[record.model].tokens += record.totalTokens;
    });

    return stats;
  }

  /**
   * 转换为RequestLog格式以便与现有系统集成
   */
  convertToRequestLog(record: TokenConsumptionRecord): Partial<RequestLog> {
    return {
      id: record.requestId,
      timestamp: record.timestamp,
      model: record.model,
      total_tokens: record.totalTokens,
      prompt_tokens: record.inputTokens,
      completion_tokens: record.outputTokens,
      cached_prompt_tokens: record.cachedPromptTokens,
      cached_completion_tokens: record.cacheReadTokens,
      is_success: true,
      duration_ms: 0, // 需要从响应时间计算
      request_type: "final" as const,
      is_stream: true
    };
  }

  /**
   * 清除所有记录
   */
  clearRecords(): void {
    this.tokenRecords = [];
    this.saveToStorage();
  }

  /**
   * 导出记录为CSV格式
   */
  exportToCsv(): string {
    const headers = [
      "timestamp", "requestId", "model", "inputTokens", "outputTokens",
      "totalTokens", "cachedTokens", "serviceTier"
    ];

    const csvContent = [
      headers.join(","),
      ...this.tokenRecords.map(record => [
        record.timestamp,
        record.requestId,
        record.model,
        record.inputTokens,
        record.outputTokens,
        record.totalTokens,
        record.cachedPromptTokens || 0,
        record.serviceTier || ""
      ].join(","))
    ].join("\n");

    return csvContent;
  }

  /**
   * 从localStorage加载数据
   */
  private loadFromStorage(): void {
    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        this.tokenRecords = JSON.parse(stored);
      }
    } catch (error) {
      console.error("Failed to load token records from storage:", error);
      this.tokenRecords = [];
    }
  }

  /**
   * 保存到localStorage
   */
  private saveToStorage(): void {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.tokenRecords));
    } catch (error) {
      console.error("Failed to save token records to storage:", error);
    }
  }

  constructor() {
    this.loadFromStorage();
  }
}

// 创建单例实例
export const claudeTokenTracker = new ClaudeTokenTracker();

// 工具函数：从Claude响应示例中提取token信息
export function extractTokensFromExample(responseText: string): TokenConsumptionRecord | null {
  return claudeTokenTracker.parseStreamingResponse(responseText);
}