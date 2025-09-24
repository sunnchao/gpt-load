import type { ApiResponse, LogFilter, LogsResponse, RequestLog } from "@/types/models";
import http from "@/utils/http";
import { calculateTokenCost, formatTokenCount as formatTokenCountUtil } from "@/utils/token-cost";

export interface TokenFilter {
  start_time?: string;
  end_time?: string;
  model?: string;
  min_tokens?: number;
  max_tokens?: number;
  group_name?: string;
}

export interface TokenStats {
  totalTokens: number;
  totalInputTokens: number;
  totalOutputTokens: number;
  totalCachedTokens: number;
  recordCount: number;
  modelBreakdown: Record<string, { count: number; tokens: number }>;
}

// 将 RequestLog 转换为 Token 记录格式
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
  reasoningTokens?: number;
  audioTokens?: number;
  imageTokens?: number;
  serviceTier?: string;
  groupName?: string;
  keyValue?: string;
  duration?: number;
  isSuccess: boolean;
}

/**
 * Claude Token Usage API
 * 基于后端 /logs 接口获取token消费数据
 */
export const claudeTokenApi = {
  /**
   * 将 RequestLog 转换为 TokenConsumptionRecord
   */
  convertLogToTokenRecord: (log: any): TokenConsumptionRecord => {
    return {
      timestamp: log.timestamp,
      requestId: log.id,
      model: log.model,
      inputTokens: log.prompt_tokens || 0,
      outputTokens: log.completion_tokens || 0,
      totalTokens: log.total_tokens || 0,
      cacheCreationTokens: log.cached_prompt_tokens,
      cacheReadTokens: log.cached_completion_tokens,
      cachedPromptTokens: (log.cached_prompt_tokens || 0) + (log.cached_completion_tokens || 0),
      reasoningTokens: log.reasoning_tokens,
      audioTokens: log.audio_tokens,
      imageTokens: log.image_tokens,
      serviceTier: undefined, // 后端暂无此字段
      groupName: log.group_name,
      keyValue: log.key_value,
      duration: log.duration_ms,
      isSuccess: log.is_success,
    };
  },

  /**
   * 获取token消费统计
   */
  getTokenStats: async (filter?: TokenFilter): Promise<ApiResponse<TokenStats>> => {
    try {
      // 构建查询参数，获取所有符合条件的记录用于统计
      const logFilter: LogFilter = {
        page: 1,
        page_size: 10000, // 获取大量数据用于统计
        start_time: filter?.start_time,
        end_time: filter?.end_time,
        model: filter?.model,
        group_name: filter?.group_name,
        is_success: true, // 只统计成功的请求
      };

      const response = await http.get<ApiResponse<LogsResponse>>("/logs", { params: logFilter });

      if (response.data.code !== 0) {
        throw new Error(response.data.message);
      }

      const logs = response.data.data?.items || [];

      // 过滤有token数据的记录
      const tokenLogs = logs.filter(
        (log: any) =>
          (log.total_tokens && log.total_tokens > 0) ||
          (log.prompt_tokens && log.prompt_tokens > 0) ||
          (log.completion_tokens && log.completion_tokens > 0)
      );

      // 应用token数量过滤
      const filteredLogs = tokenLogs.filter((log: any) => {
        if (filter?.min_tokens && (log.total_tokens || 0) < filter.min_tokens) {
          return false;
        }
        if (filter?.max_tokens && (log.total_tokens || 0) > filter.max_tokens) {
          return false;
        }
        return true;
      });

      // 计算统计信息
      const stats: TokenStats = {
        totalTokens: 0,
        totalInputTokens: 0,
        totalOutputTokens: 0,
        totalCachedTokens: 0,
        recordCount: filteredLogs.length,
        modelBreakdown: {},
      };

      filteredLogs.forEach(log => {
        const totalTokens = log.total_tokens || 0;
        const inputTokens = log.prompt_tokens || 0;
        const outputTokens = log.completion_tokens || 0;
        const cachedTokens = (log.cached_prompt_tokens || 0) + (log.cached_completion_tokens || 0);

        stats.totalTokens += totalTokens;
        stats.totalInputTokens += inputTokens;
        stats.totalOutputTokens += outputTokens;
        stats.totalCachedTokens += cachedTokens;

        // 模型统计
        if (!stats.modelBreakdown[log.model]) {
          stats.modelBreakdown[log.model] = { count: 0, tokens: 0 };
        }
        stats.modelBreakdown[log.model].count++;
        stats.modelBreakdown[log.model].tokens += totalTokens;
      });

      return {
        code: 0,
        message: "Success",
        data: stats,
      };
    } catch (error) {
      console.error("Failed to get token stats:", error);
      return {
        code: -1,
        message: `Failed to get token stats: ${error}`,
        data: {
          totalTokens: 0,
          totalInputTokens: 0,
          totalOutputTokens: 0,
          totalCachedTokens: 0,
          recordCount: 0,
          modelBreakdown: {},
        },
      };
    }
  },

  /**
   * 获取token消费记录列表
   */
  getTokenRecords: async (
    filter?: TokenFilter,
    page: number = 1,
    pageSize: number = 15
  ): Promise<ApiResponse<{ records: TokenConsumptionRecord[]; pagination: any }>> => {
    try {
      const logFilter: LogFilter = {
        page,
        page_size: pageSize,
        start_time: filter?.start_time,
        end_time: filter?.end_time,
        model: filter?.model,
        group_name: filter?.group_name,
        is_success: true, // 只获取成功的请求
      };

      const response = await http.get<ApiResponse<LogsResponse>>("/logs", { params: logFilter });

      if (response.data.code !== 0) {
        throw new Error(response.data.message);
      }

      const logs = response.data.data?.items || [];

      // 过滤有token数据的记录并转换格式
      let tokenRecords = logs
        .filter(
          (log: RequestLog) =>
            (log.total_tokens && log.total_tokens > 0) ||
            (log.prompt_tokens && log.prompt_tokens > 0) ||
            (log.completion_tokens && log.completion_tokens > 0)
        )
        .map((log: RequestLog) => claudeTokenApi.convertLogToTokenRecord(log));

      // 应用token数量过滤
      if (filter?.min_tokens || filter?.max_tokens) {
        tokenRecords = tokenRecords.filter((record: TokenConsumptionRecord) => {
          if (filter.min_tokens && record.totalTokens < filter.min_tokens) {
            return false;
          }
          if (filter.max_tokens && record.totalTokens > filter.max_tokens) {
            return false;
          }
          return true;
        });
      }

      return {
        code: 0,
        message: "Success",
        data: {
          records: tokenRecords,
          pagination: response.data.data?.pagination,
        },
      };
    } catch (error) {
      console.error("Failed to get token records:", error);
      return {
        code: -1,
        message: `Failed to get token records: ${error}`,
        data: {
          records: [],
          pagination: { page: 1, page_size: pageSize, total_items: 0, total_pages: 0 },
        },
      };
    }
  },

  /**
   * 导出token消费记录
   */
  exportTokenRecords: (filter?: TokenFilter): void => {
    try {
      // 使用现有的日志导出功能
      const logFilter: Omit<LogFilter, "page" | "page_size"> = {
        start_time: filter?.start_time,
        end_time: filter?.end_time,
        model: filter?.model,
        group_name: filter?.group_name,
        is_success: true,
      };

      // 导出日志（包含token信息）
      const authKey = localStorage.getItem("authKey");
      if (!authKey) {
        window.$message?.error("未找到认证密钥");
        return;
      }

      const queryParams = new URLSearchParams(
        Object.entries(logFilter).reduce(
          (acc, [key, value]) => {
            if (value !== undefined && value !== null && value !== "") {
              acc[key] = String(value);
            }
            return acc;
          },
          {} as Record<string, string>
        )
      );
      queryParams.append("key", authKey);

      const url = `${http.defaults.baseURL}/logs/export?${queryParams.toString()}`;

      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", `claude-token-usage-${Date.now()}.csv`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      window.$message?.success("token使用记录导出成功");
    } catch (error) {
      console.error("Failed to export token records:", error);
      window.$message?.error("导出失败");
    }
  },

  /**
   * 清除所有token记录（实际上无法清除后端日志，此功能保留以兼容现有接口）
   */
  clearTokenRecords: (): Promise<ApiResponse<boolean>> => {
    // 由于使用后端数据，无法直接清除记录
    window.$message?.warning("无法清除后端日志记录，请联系管理员");
    return Promise.resolve({
      code: -1,
      message: "Cannot clear backend logs",
      data: false,
    });
  },
};

// 重新导出工具函数以保持向后兼容
export const formatTokenCount = formatTokenCountUtil;
export const estimateCost = calculateTokenCost;
