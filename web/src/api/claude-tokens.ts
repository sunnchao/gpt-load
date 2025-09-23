import type { ApiResponse } from "@/types/models";
import http from "@/utils/http";
import { claudeTokenTracker, type TokenConsumptionRecord } from "@/services/claude-token-tracker";

export interface TokenFilter {
  start_time?: string;
  end_time?: string;
  model?: string;
  min_tokens?: number;
  max_tokens?: number;
}

export interface TokenStats {
  totalTokens: number;
  totalInputTokens: number;
  totalOutputTokens: number;
  totalCachedTokens: number;
  recordCount: number;
  modelBreakdown: Record<string, { count: number; tokens: number }>;
}

/**
 * Claude Token Usage API
 * 管理Claude API token消费数据的API接口
 */
export const claudeTokenApi = {
  /**
   * 处理Claude API流式响应并记录token使用量
   */
  processClaudeResponse: async (responseText: string): Promise<TokenConsumptionRecord | null> => {
    try {
      const record = claudeTokenTracker.processAndStore(responseText);
      if (record) {
        // 可选：发送到后端服务器持久化存储
        // await http.post("/tokens/claude", record);
      }
      return record;
    } catch (error) {
      console.error("Failed to process Claude response:", error);
      return null;
    }
  },

  /**
   * 获取token消费统计
   */
  getTokenStats: (): Promise<ApiResponse<TokenStats>> => {
    try {
      const stats = claudeTokenTracker.getUsageStats();
      return Promise.resolve({
        code: 0,
        message: "Success",
        data: stats
      });
    } catch (error) {
      return Promise.resolve({
        code: -1,
        message: `Failed to get token stats: ${error}`,
        data: {
          totalTokens: 0,
          totalInputTokens: 0,
          totalOutputTokens: 0,
          totalCachedTokens: 0,
          recordCount: 0,
          modelBreakdown: {}
        }
      });
    }
  },

  /**
   * 获取token消费记录列表
   */
  getTokenRecords: (filter?: TokenFilter): Promise<ApiResponse<TokenConsumptionRecord[]>> => {
    try {
      let records = claudeTokenTracker.getTokenRecords();

      // 应用过滤器
      if (filter) {
        records = records.filter(record => {
          // 时间范围过滤
          if (filter.start_time && record.timestamp < filter.start_time) return false;
          if (filter.end_time && record.timestamp > filter.end_time) return false;

          // 模型过滤
          if (filter.model && !record.model.includes(filter.model)) return false;

          // token数量范围过滤
          if (filter.min_tokens && record.totalTokens < filter.min_tokens) return false;
          if (filter.max_tokens && record.totalTokens > filter.max_tokens) return false;

          return true;
        });
      }

      // 按时间倒序排列
      records.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

      return Promise.resolve({
        code: 0,
        message: "Success",
        data: records
      });
    } catch (error) {
      return Promise.resolve({
        code: -1,
        message: `Failed to get token records: ${error}`,
        data: []
      });
    }
  },

  /**
   * 导出token消费记录
   */
  exportTokenRecords: (filter?: TokenFilter): void => {
    try {
      // TODO: 将来可以基于filter参数过滤导出的记录
      console.log("Export filter:", filter);

      const csvContent = claudeTokenTracker.exportToCsv();

      // 创建下载链接
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement("a");
      const url = URL.createObjectURL(blob);

      link.setAttribute("href", url);
      link.setAttribute("download", `claude-token-usage-${Date.now()}.csv`);
      link.style.visibility = 'hidden';

      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error("Failed to export token records:", error);
      window.$message?.error("导出失败");
    }
  },

  /**
   * 清除所有token记录
   */
  clearTokenRecords: (): Promise<ApiResponse<boolean>> => {
    try {
      claudeTokenTracker.clearRecords();
      return Promise.resolve({
        code: 0,
        message: "Records cleared successfully",
        data: true
      });
    } catch (error) {
      return Promise.resolve({
        code: -1,
        message: `Failed to clear records: ${error}`,
        data: false
      });
    }
  },

  /**
   * 手动解析Claude响应示例（用于测试）
   */
  parseResponseExample: (responseText: string): Promise<ApiResponse<TokenConsumptionRecord | null>> => {
    try {
      const record = claudeTokenTracker.parseStreamingResponse(responseText);
      return Promise.resolve({
        code: 0,
        message: "Parsing successful",
        data: record
      });
    } catch (error) {
      return Promise.resolve({
        code: -1,
        message: `Parsing failed: ${error}`,
        data: null
      });
    }
  }
};

/**
 * HTTP拦截器：自动检测和处理Claude API响应
 */
export const setupClaudeTokenInterceptor = () => {
  // 响应拦截器
  http.interceptors.response.use(
    (response) => {
      // 检查是否为Claude API响应（基于URL或响应头）
      const url = response.config.url || "";
      const isClaudeApi = url.includes("claude") || url.includes("anthropic");

      if (isClaudeApi && response.data) {
        try {
          // 尝试解析并存储token使用量
          const responseText = typeof response.data === 'string' ? response.data : JSON.stringify(response.data);
          claudeTokenApi.processClaudeResponse(responseText);
        } catch (error) {
          console.warn("Failed to process Claude response for token tracking:", error);
        }
      }

      return response;
    },
    (error) => {
      return Promise.reject(error);
    }
  );
};

// 工具函数：格式化token数量显示
export const formatTokenCount = (count: number): string => {
  if (count < 1000) return count.toString();
  if (count < 1000000) return (count / 1000).toFixed(1) + 'K';
  return (count / 1000000).toFixed(1) + 'M';
};

// 工具函数：计算成本估算（基于Claude定价）
export const estimateCost = (inputTokens: number, outputTokens: number, model: string): number => {
  // Claude定价（示例，实际价格请参考官方文档）
  const pricing: Record<string, { input: number; output: number }> = {
    "claude-sonnet-4": { input: 0.003, output: 0.015 }, // per 1K tokens
    "claude-haiku-3": { input: 0.00025, output: 0.00125 },
    "claude-opus-3": { input: 0.015, output: 0.075 }
  };

  // 查找匹配的模型价格
  let modelPricing = pricing["claude-sonnet-4"]; // 默认价格
  for (const [modelName, price] of Object.entries(pricing)) {
    if (model.includes(modelName)) {
      modelPricing = price;
      break;
    }
  }

  const inputCost = (inputTokens / 1000) * modelPricing.input;
  const outputCost = (outputTokens / 1000) * modelPricing.output;

  return inputCost + outputCost;
};