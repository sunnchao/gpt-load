<script setup lang="ts">
import {
  claudeTokenApi,
  estimateCost,
  formatTokenCount,
  type TokenFilter,
  type TokenStats,
} from "@/api/claude-tokens";
import type { TokenConsumptionRecord } from "@/services/claude-token-tracker";
import { copy } from "@/utils/clipboard";
import {
  CloudDownloadOutline,
  CodeSlashOutline,
  CopyOutline,
  DocumentTextOutline,
  PlayOutline,
  RefreshOutline,
  StatsChartOutline,
  TrashOutline,
} from "@vicons/ionicons5";
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NEllipsis,
  NGrid,
  NGridItem,
  NIcon,
  NInput,
  NInputGroup,
  NInputGroupLabel,
  NModal,
  NPopconfirm,
  NSpace,
  NStatistic,
  NTag,
  NTooltip,
  useMessage,
} from "naive-ui";
import { computed, h, onMounted, reactive, ref, type VNodeChild } from "vue";

const message = useMessage();

// Data
const loading = ref(false);
const tokenRecords = ref<TokenConsumptionRecord[]>([]);
const tokenStats = ref<TokenStats>({
  totalTokens: 0,
  totalInputTokens: 0,
  totalOutputTokens: 0,
  totalCachedTokens: 0,
  recordCount: 0,
  modelBreakdown: {},
});

// Modal for viewing response details
const showDetailModal = ref(false);
const selectedRecord = ref<TokenConsumptionRecord | null>(null);

// Test modal
const showTestModal = ref(false);
const testResponseText = ref("");

// Filters
const filters = reactive({
  start_time: null as number | null,
  end_time: null as number | null,
  model: "",
  min_tokens: "" as string,
  max_tokens: "" as string,
});

// Load data
const loadTokenData = async () => {
  loading.value = true;
  try {
    const [statsRes, recordsRes] = await Promise.all([
      claudeTokenApi.getTokenStats(),
      claudeTokenApi.getTokenRecords({
        start_time: filters.start_time ? new Date(filters.start_time).toISOString() : undefined,
        end_time: filters.end_time ? new Date(filters.end_time).toISOString() : undefined,
        model: filters.model || undefined,
        min_tokens: filters.min_tokens ? parseInt(filters.min_tokens) : undefined,
        max_tokens: filters.max_tokens ? parseInt(filters.max_tokens) : undefined,
      }),
    ]);

    if (statsRes.code === 0) {
      tokenStats.value = statsRes.data;
    }

    if (recordsRes.code === 0) {
      tokenRecords.value = recordsRes.data;
    }
  } catch (error) {
    console.error("Failed to load token data:", error);
    message.error("加载token数据失败");
  } finally {
    loading.value = false;
  }
};

// Format date time
const formatDateTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString("zh-CN", { hour12: false }).replace(/\//g, "-");
};

// View response details
const viewResponseDetails = (record: TokenConsumptionRecord) => {
  selectedRecord.value = record;
  showDetailModal.value = true;
};

// Copy content
const copyContent = async (content: string, type: string) => {
  const success = await copy(content);
  if (success) {
    message.success(`${type} 已复制到剪贴板`);
  } else {
    message.error(`复制 ${type} 失败`);
  }
};

// Export records
const exportRecords = () => {
  const apiFilter: TokenFilter = {
    start_time: filters.start_time ? new Date(filters.start_time).toISOString() : undefined,
    end_time: filters.end_time ? new Date(filters.end_time).toISOString() : undefined,
    model: filters.model || undefined,
    min_tokens: filters.min_tokens ? parseInt(filters.min_tokens) : undefined,
    max_tokens: filters.max_tokens ? parseInt(filters.max_tokens) : undefined,
  };

  claudeTokenApi.exportTokenRecords(apiFilter);
  message.success("token使用记录导出成功");
};

// Clear all records
const clearAllRecords = async () => {
  try {
    const res = await claudeTokenApi.clearTokenRecords();
    if (res.code === 0) {
      message.success("所有记录已清除");
      await loadTokenData();
    } else {
      message.error(res.message);
    }
  } catch (error) {
    message.error("清除记录失败");
  }
};

// Test parsing
const testParsing = async () => {
  if (!testResponseText.value.trim()) {
    message.warning("请输入Claude API响应数据");
    return;
  }

  try {
    const res = await claudeTokenApi.parseResponseExample(testResponseText.value);
    if (res.code === 0 && res.data) {
      message.success("解析成功！token信息已提取");
      showTestModal.value = false;
      await loadTokenData(); // 重新加载数据
    } else {
      message.error(res.message || "解析失败");
    }
  } catch (error) {
    message.error("解析过程中发生错误");
  }
};

// Handle search
const handleSearch = () => {
  loadTokenData();
};

// Reset filters
const resetFilters = () => {
  Object.assign(filters, {
    start_time: null,
    end_time: null,
    model: "",
    min_tokens: "",
    max_tokens: "",
  });
  loadTokenData();
};

// Table columns
const columns = computed(() => [
  {
    title: "时间",
    key: "timestamp",
    width: 160,
    render: (row: TokenConsumptionRecord) => formatDateTime(row.timestamp),
  },
  {
    title: "请求ID",
    key: "requestId",
    width: 200,
    render: (row: TokenConsumptionRecord) =>
      h(NEllipsis, { style: "max-width: 180px" }, { default: () => row.requestId }),
  },
  {
    title: "模型",
    key: "model",
    width: 200,
    render: (row: TokenConsumptionRecord) =>
      h(NTag, { type: "info", size: "small" }, { default: () => row.model }),
  },
  {
    title: "Token使用量",
    key: "tokens",
    width: 280,
    render: (row: TokenConsumptionRecord) => {
      const parts: VNodeChild[] = [];

      // 总计
      parts.push(
        h(
          "div",
          { style: "font-weight: bold; color: var(--primary-color); margin-bottom: 2px" },
          `总计: ${formatTokenCount(row.totalTokens)}`
        )
      );

      // 输入输出
      parts.push(
        h(
          "div",
          { style: "color: var(--text-secondary); font-size: 11px; margin-bottom: 2px" },
          `输入: ${formatTokenCount(row.inputTokens)} | 输出: ${formatTokenCount(row.outputTokens)}`
        )
      );

      // 缓存相关信息
      const cacheInfo: string[] = [];
      if (row.cacheCreationTokens) {
        cacheInfo.push(`创建: ${formatTokenCount(row.cacheCreationTokens)}`);
      }
      if (row.cacheReadTokens) {
        cacheInfo.push(`读取: ${formatTokenCount(row.cacheReadTokens)}`);
      }
      if (row.ephemeral5mTokens) {
        cacheInfo.push(`5m临时: ${formatTokenCount(row.ephemeral5mTokens)}`);
      }
      if (row.ephemeral1hTokens) {
        cacheInfo.push(`1h临时: ${formatTokenCount(row.ephemeral1hTokens)}`);
      }

      if (cacheInfo.length > 0) {
        parts.push(
          h(
            "div",
            { style: "color: var(--info-color); font-size: 10px" },
            `缓存: ${cacheInfo.join(' | ')}`
          )
        );
      }

      return h("div", { style: "line-height: 1.4" }, parts);
    },
  },
  {
    title: "预估成本",
    key: "cost",
    width: 100,
    render: (row: TokenConsumptionRecord) => {
      const cost = estimateCost(row.inputTokens, row.outputTokens, row.model);
      return h(
        "span",
        { style: "color: var(--warning-color); font-weight: 500" },
        `$${cost.toFixed(4)}`
      );
    },
  },
  {
    title: "服务层级",
    key: "serviceTier",
    width: 100,
    render: (row: TokenConsumptionRecord) =>
      row.serviceTier
        ? h(NTag, { type: "success", size: "small" }, { default: () => row.serviceTier })
        : "-",
  },
  {
    title: "操作",
    key: "actions",
    width: 120,
    fixed: "right" as const,
    render: (row: TokenConsumptionRecord) =>
      h(NSpace, { size: "small" }, () => [
        h(
          NButton,
          {
            size: "small",
            type: "primary",
            ghost: true,
            onClick: () => viewResponseDetails(row),
          },
          {
            icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }),
            default: () => "详情",
          }
        ),
      ]),
  },
]);

// Computed scroll width
const scrollX = computed(() => 1160);

onMounted(() => {
  loadTokenData();
});
</script>

<template>
  <div class="claude-token-container">
    <n-space vertical size="large">
      <!-- 统计卡片 -->
      <n-grid x-gap="16" y-gap="16" cols="1 s:2 m:4" responsive="screen">
        <n-grid-item>
          <n-card>
            <n-statistic label="总Token消费" :value="formatTokenCount(tokenStats.totalTokens)">
              <template #prefix>
                <n-icon :component="StatsChartOutline" />
              </template>
            </n-statistic>
          </n-card>
        </n-grid-item>
        <n-grid-item>
          <n-card>
            <n-statistic label="输入Token" :value="formatTokenCount(tokenStats.totalInputTokens)" />
          </n-card>
        </n-grid-item>
        <n-grid-item>
          <n-card>
            <n-statistic
              label="输出Token"
              :value="formatTokenCount(tokenStats.totalOutputTokens)"
            />
          </n-card>
        </n-grid-item>
        <n-grid-item>
          <n-card>
            <n-statistic label="请求次数" :value="tokenStats.recordCount" />
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 工具栏 -->
      <n-card title="Claude Token 使用记录" size="small">
        <template #header-extra>
          <n-space>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button size="small" @click="showTestModal = true">
                  <template #icon>
                    <n-icon :component="PlayOutline" />
                  </template>
                  测试解析
                </n-button>
              </template>
              测试解析Claude API响应
            </n-tooltip>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button size="small" @click="loadTokenData">
                  <template #icon>
                    <n-icon :component="RefreshOutline" />
                  </template>
                </n-button>
              </template>
              刷新数据
            </n-tooltip>
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button size="small" @click="exportRecords">
                  <template #icon>
                    <n-icon :component="CloudDownloadOutline" />
                  </template>
                </n-button>
              </template>
              导出记录
            </n-tooltip>
            <n-popconfirm @positive-click="clearAllRecords">
              <template #trigger>
                <n-button size="small" type="error" secondary>
                  <template #icon>
                    <n-icon :component="TrashOutline" />
                  </template>
                </n-button>
              </template>
              确认清除所有token使用记录？此操作不可撤销。
            </n-popconfirm>
          </n-space>
        </template>

        <!-- 过滤器 -->
        <div style="margin-bottom: 16px">
          <n-space>
            <n-input-group>
              <n-input-group-label>模型</n-input-group-label>
              <n-input
                v-model:value="filters.model"
                placeholder="输入模型名称过滤"
                style="width: 200px"
                clearable
                @keyup.enter="handleSearch"
              />
            </n-input-group>
            <n-input-group>
              <n-input-group-label>最小Token</n-input-group-label>
              <n-input
                v-model:value="filters.min_tokens"
                placeholder="0"
                style="width: 120px"
                clearable
                @keyup.enter="handleSearch"
              />
            </n-input-group>
            <n-input-group>
              <n-input-group-label>最大Token</n-input-group-label>
              <n-input
                v-model:value="filters.max_tokens"
                placeholder="不限"
                style="width: 120px"
                clearable
                @keyup.enter="handleSearch"
              />
            </n-input-group>
            <n-date-picker
              v-model:value="filters.start_time"
              type="datetime"
              placeholder="开始时间"
              clearable
              style="width: 180px"
              @update:value="handleSearch"
            />
            <n-date-picker
              v-model:value="filters.end_time"
              type="datetime"
              placeholder="结束时间"
              clearable
              style="width: 180px"
              @update:value="handleSearch"
            />
            <n-button @click="resetFilters">重置</n-button>
          </n-space>
        </div>

        <!-- 数据表格 -->
        <n-data-table
          :columns="columns"
          :data="tokenRecords"
          :loading="loading"
          :scroll-x="scrollX"
          size="small"
          remote
        />
      </n-card>
    </n-space>

    <!-- 详情模态框 -->
    <n-modal
      v-model:show="showDetailModal"
      preset="card"
      style="width: 800px"
      title="Token使用详情"
    >
      <div v-if="selectedRecord" style="max-height: 60vh; overflow-y: auto">
        <n-space vertical size="small">
          <n-card title="基本信息" size="small">
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">请求ID:</span>
                <span class="detail-value">{{ selectedRecord.requestId }}</span>
                <n-button size="tiny" text @click="copyContent(selectedRecord.requestId, '请求ID')">
                  <template #icon>
                    <n-icon :component="CopyOutline" />
                  </template>
                </n-button>
              </div>
              <div class="detail-item">
                <span class="detail-label">时间:</span>
                <span class="detail-value">{{ formatDateTime(selectedRecord.timestamp) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">模型:</span>
                <span class="detail-value">{{ selectedRecord.model }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">输入Token:</span>
                <span class="detail-value">{{ formatTokenCount(selectedRecord.inputTokens) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">输出Token:</span>
                <span class="detail-value">
                  {{ formatTokenCount(selectedRecord.outputTokens) }}
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">总计Token:</span>
                <span class="detail-value">{{ formatTokenCount(selectedRecord.totalTokens) }}</span>
              </div>
              <div v-if="selectedRecord.cacheCreationTokens" class="detail-item">
                <span class="detail-label">缓存创建Token:</span>
                <span class="detail-value">
                  {{ formatTokenCount(selectedRecord.cacheCreationTokens) }}
                </span>
              </div>
              <div v-if="selectedRecord.cacheReadTokens" class="detail-item">
                <span class="detail-label">缓存读取Token:</span>
                <span class="detail-value">
                  {{ formatTokenCount(selectedRecord.cacheReadTokens) }}
                </span>
              </div>
              <div v-if="selectedRecord.ephemeral5mTokens" class="detail-item">
                <span class="detail-label">5分钟临时缓存:</span>
                <span class="detail-value">
                  {{ formatTokenCount(selectedRecord.ephemeral5mTokens) }}
                </span>
              </div>
              <div v-if="selectedRecord.ephemeral1hTokens" class="detail-item">
                <span class="detail-label">1小时临时缓存:</span>
                <span class="detail-value">
                  {{ formatTokenCount(selectedRecord.ephemeral1hTokens) }}
                </span>
              </div>
              <div v-if="selectedRecord.serviceTier" class="detail-item">
                <span class="detail-label">服务层级:</span>
                <span class="detail-value">{{ selectedRecord.serviceTier }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">预估成本:</span>
                <span class="detail-value cost">
                  ${{
                    estimateCost(
                      selectedRecord.inputTokens,
                      selectedRecord.outputTokens,
                      selectedRecord.model
                    ).toFixed(4)
                  }}
                </span>
              </div>
            </div>
          </n-card>

          <n-card v-if="selectedRecord.rawResponse" title="原始响应" size="small">
            <template #header-extra>
              <n-button
                size="tiny"
                text
                @click="copyContent(selectedRecord.rawResponse || '', '原始响应')"
              >
                <template #icon>
                  <n-icon :component="CopyOutline" />
                </template>
              </n-button>
            </template>
            <div class="response-content">
              {{ selectedRecord.rawResponse }}
            </div>
          </n-card>
        </n-space>
      </div>
    </n-modal>

    <!-- 测试解析模态框 -->
    <n-modal
      v-model:show="showTestModal"
      preset="card"
      style="width: 900px"
      title="测试Claude响应解析"
    >
      <n-space vertical>
        <div>
          <p style="color: var(--text-secondary); font-size: 13px; margin-bottom: 8px">
            请粘贴Claude API的流式响应数据，系统将自动解析并提取token使用量信息：
          </p>
          <n-input
            v-model:value="testResponseText"
            type="textarea"
            placeholder="event: message_start&#10;data: {...}&#10;&#10;event: content_block_delta&#10;data: {...}&#10;&#10;..."
            :autosize="{ minRows: 10, maxRows: 20 }"
          />
        </div>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showTestModal = false">取消</n-button>
          <n-button type="primary" @click="testParsing">
            <template #icon>
              <n-icon :component="CodeSlashOutline" />
            </template>
            解析并存储
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.claude-token-container {
  padding: 0;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-label {
  font-weight: 500;
  color: var(--text-secondary);
  min-width: 80px;
  flex-shrink: 0;
}

.detail-value {
  color: var(--text-primary);
  flex: 1;
}

.detail-value.cost {
  color: var(--warning-color);
  font-weight: 600;
}

.response-content {
  background: var(--code-block-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.4;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
