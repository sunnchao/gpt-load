/**
 * Token 成本计算工具函数
 * 从 claude-tokens.ts 中抽取出来，作为独立的工具函数
 */

// Claude定价配置（人民币计费，per 1K tokens）
export const TOKEN_PRICING = {
  "claude-sonnet-4": {
    input: 0.003, // ¥3 / M Tokens = ¥0.003 / K Tokens
    output: 0.015, // ¥15 / M Tokens = ¥0.015 / K Tokens
    cacheWrite: 0.00375, // ¥3.75 / M Tokens = ¥0.00375 / K Tokens
    cacheHit: 0.0003, // ¥0.3 / M Tokens = ¥0.0003 / K Tokens
  },
  "claude-3-5-haiku-20241022": {
    input: 0.001, // ¥1 / M Tokens = ¥0.001 / K Tokens
    output: 0.005, // ¥5 / M Tokens = ¥0.005 / K Tokens
    cacheWrite: 0.00125, // ¥1.25 / M Tokens = ¥0.00125 / K Tokens
    cacheHit: 0.0001, // ¥0.1 / M Tokens = ¥0.0001 / K Tokens
  },
  "claude-opus-4-20250514": {
    input: 0.015, // ¥15 / M Tokens = ¥0.015 / K Tokens
    output: 0.075, // ¥75 / M Tokens = ¥0.075 / K Tokens
    cacheWrite: 0.01875, // ¥18.75 / M Tokens = ¥0.01875 / K Tokens
    cacheHit: 0.0015, // ¥1.5 / M Tokens = ¥0.0015 / K Tokens
  },
  "claude-opus-4-1-20250805": {
    input: 0.015, // ¥15 / M Tokens = ¥0.015 / K Tokens
    output: 0.075, // ¥75 / M Tokens = ¥0.075 / K Tokens
    cacheWrite: 0.01875, // ¥18.75 / M Tokens = ¥0.01875 / K Tokens
    cacheHit: 0.0015, // ¥1.5 / M Tokens = ¥0.0015 / K Tokens
  },
  "claude-haiku-3": {
    input: 0.00025,
    output: 0.00125,
    cacheWrite: 0.0003125,
    cacheHit: 0.000025,
  },
  "claude-opus-3": {
    input: 0.015,
    output: 0.075,
    cacheWrite: 0.01875,
    cacheHit: 0.0015,
  },
} as const;

export type ModelPricing = {
  input: number;
  output: number;
  cacheWrite: number;
  cacheHit: number;
};

/**
 * 根据模型名称获取定价信息
 */
export function getModelPricing(model: string): ModelPricing {
  // 查找匹配的模型价格
  for (const [modelName, pricing] of Object.entries(TOKEN_PRICING)) {
    if (model.includes(modelName)) {
      return pricing;
    }
  }
  // 默认使用 Claude Sonnet 4 价格
  return TOKEN_PRICING["claude-sonnet-4"];
}

/**
 * 计算 Token 成本
 * @param inputTokens 输入token数量
 * @param outputTokens 输出token数量
 * @param model 模型名称
 * @param cacheCreationTokens 缓存创建token数量（可选）
 * @param cacheReadTokens 缓存读取token数量（可选）
 * @returns 计算出的成本（人民币）
 */
export function calculateTokenCost(
  inputTokens: number,
  outputTokens: number,
  model: string,
  cacheCreationTokens?: number,
  cacheReadTokens?: number
): number {
  const pricing = getModelPricing(model);

  const inputCost = (inputTokens / 1000) * pricing.input;
  const outputCost = (outputTokens / 1000) * pricing.output;
  const cacheWriteCost = cacheCreationTokens
    ? (cacheCreationTokens / 1000) * pricing.cacheWrite
    : 0;
  const cacheHitCost = cacheReadTokens ? (cacheReadTokens / 1000) * pricing.cacheHit : 0;

  return inputCost + outputCost + cacheWriteCost + cacheHitCost;
}

/**
 * 格式化成本显示
 * @param cost 成本金额
 * @param currency 货币符号，默认为 ¥
 * @param precision 小数位数，默认为 4
 */
export function formatCost(cost: number, currency: string = "¥", precision: number = 4): string {
  return `${currency}${cost.toFixed(precision)}`;
}

/**
 * 格式化Token数量显示
 */
export function formatTokenCount(count: number): string {
  if (count < 1000) {
    return count.toString();
  }
  if (count < 1000000) {
    return `${(count / 1000).toFixed(1)}K`;
  }
  return `${(count / 1000000).toFixed(1)}M`;
}
