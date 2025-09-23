import { claudeTokenTracker } from "@/services/claude-token-tracker";

// 你提供的Claude API响应示例数据
const exampleClaudeResponse = `event: message_start
data: {"type":"message_start","message":{"id":"msg_01U2Nmy9ZeAnB8h2GGNhge1R","type":"message","role":"assistant","model":"claude-sonnet-4-20250514","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":4,"cache_creation_input_tokens":3592,"cache_read_input_tokens":11357,"cache_creation":{"ephemeral_5m_input_tokens":3592,"ephemeral_1h_input_tokens":0},"output_tokens":1,"service_tier":"standard"}}             }

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}      }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hi"}      }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"! How can I help you with your"}            }

event: ping
data: {"type": "ping"}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" code today?"}              }

event: ping
data: {"type": "ping"}

event: content_block_stop
data: {"type":"content_block_stop","index":0            }

event: ping
data: {"type": "ping"}

event: ping
data: {"type": "ping"}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"input_tokens":4,"cache_creation_input_tokens":3592,"cache_read_input_tokens":11357,"output_tokens":15}          }

event: message_stop
data: {"type":"message_stop" }`;

/**
 * 测试Claude token追踪功能
 */
export const testClaudeTokenTracking = () => {
  console.log("开始测试Claude token追踪功能...");

  // 测试解析功能
  const record = claudeTokenTracker.parseStreamingResponse(exampleClaudeResponse);

  if (record) {
    console.log("✅ 解析成功！提取的token信息:");
    console.log("- 请求ID:", record.requestId);
    console.log("- 模型:", record.model);
    console.log("- 输入Tokens:", record.inputTokens);
    console.log("- 输出Tokens:", record.outputTokens);
    console.log("- 总Tokens:", record.totalTokens);
    console.log("- 缓存创建Tokens:", record.cacheCreationTokens);
    console.log("- 缓存读取Tokens:", record.cacheReadTokens);
    console.log("- 服务层级:", record.serviceTier);

    // 存储记录
    claudeTokenTracker.storeTokenRecord(record);
    console.log("✅ Token记录已存储到本地");

    // 获取统计信息
    const stats = claudeTokenTracker.getUsageStats();
    console.log("✅ 当前统计信息:");
    console.log("- 总记录数:", stats.recordCount);
    console.log("- 总Token消费:", stats.totalTokens);
    console.log("- 总输入Token:", stats.totalInputTokens);
    console.log("- 总输出Token:", stats.totalOutputTokens);
    console.log("- 模型分布:", stats.modelBreakdown);

    return record;
  } else {
    console.error("❌ 解析失败");
    return null;
  }
};

// 自动在页面加载时测试（仅开发环境）
if (import.meta.env.DEV) {
  // 延迟执行，确保页面加载完成
  setTimeout(() => {
    testClaudeTokenTracking();
  }, 1000);
}