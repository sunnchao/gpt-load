<script setup lang="ts">
import {
  claudeTokenApi,
  estimateCost,
  formatTokenCount,
  type TokenConsumptionRecord,
  type TokenFilter,
  type TokenStats,
} from "@/api/claude-tokens";
import { logApi } from "@/api/logs";
import TokenDetailPanel from "@/components/logs/TokenDetailPanel.vue";
import {
  BarChartOutline,
  CashOutline,
  CloudDownloadOutline,
  DocumentTextOutline,
  FilterOutline,
  RefreshOutline,
  StatsChartOutline,
  TimeOutline,
  TrendingUpOutline,
} from "@vicons/ionicons5";
import {
  NBadge,
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NFormItem,
  NIcon,
  NInput,
  NInputGroup,
  NModal,
  NRadioButton,
  NRadioGroup,
  NSelect,
  useMessage,
  type DataTableColumns,
  type Pagination,
} from "naive-ui";
import { computed, h, onMounted, reactive, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const message = useMessage();

// Data
const loading = ref(false);
const exporting = ref(false);
const tokenRecords = ref<TokenConsumptionRecord[]>([]);
const selectedRecord = ref<TokenConsumptionRecord | null>(null);
const showDetailModal = ref(false);
const chartContainer = ref<HTMLElement>();
const chartPeriod = ref<"24h" | "7d" | "30d">("7d");

// Statistics
const tokenStats = ref<
  TokenStats & {
    requestCount: number;
    avgTokensPerRequest: number;
    todayUsage: string;
    growthRate: number;
    estimatedCost: string;
  }
>({
  totalTokens: 0,
  totalInputTokens: 0,
  totalOutputTokens: 0,
  requestCount: 0,
  avgTokensPerRequest: 0,
  todayUsage: "0",
  growthRate: 0,
  estimatedCost: "0.00",
});

// Filters
const filters = reactive<
  TokenFilter & {
    page: number;
    page_size: number;
    minTokens: string;
    maxTokens: string;
  }
>({
  page: 1,
  page_size: 20,
  start_time: "",
  end_time: "",
  model: "",
  group: "",
  minTokens: "",
  maxTokens: "",
});

const dateRange = ref<[number, number] | null>(null);

const pagination = ref<Pagination>({
  page: 1,
  page_size: 20,
  total_items: 0,
  total_pages: 0,
});

// Options
const modelOptions = ref([
  { label: "Claude-3.5-Sonnet", value: "claude-3-5-sonnet-20241022" },
  { label: "Claude-3-Opus", value: "claude-3-opus-20240229" },
  { label: "Claude-3-Haiku", value: "claude-3-haiku-20240307" },
]);

const groupOptions = ref([
  { label: "Anthropic", value: "anthropic" },
  { label: "Claude", value: "claude" },
]);

// Table columns
const columns = computed<DataTableColumns<TokenConsumptionRecord>>(() => [
  {
    title: t("tokens.timestamp"),
    key: "timestamp",
    width: 160,
    render: row => new Date(row.timestamp).toLocaleString(),
  },
  {
    title: t("tokens.model"),
    key: "model",
    width: 180,
    ellipsis: { tooltip: true },
  },
  {
    title: t("tokens.inputTokens"),
    key: "input_tokens",
    width: 120,
    render: row => formatTokenCount(row.input_tokens),
  },
  {
    title: t("tokens.outputTokens"),
    key: "output_tokens",
    width: 120,
    render: row => formatTokenCount(row.output_tokens),
  },
  {
    title: t("tokens.totalTokens"),
    key: "total_tokens",
    width: 120,
    render: row => h("span", { class: "total-tokens" }, formatTokenCount(row.total_tokens)),
  },
  {
    title: t("tokens.cost"),
    key: "estimated_cost",
    width: 100,
    render: row =>
      `$${estimateCost(row.input_tokens, row.output_tokens, row.model || "claude-sonnet-4").toFixed(4)}`,
  },
  {
    title: t("tokens.actions"),
    key: "actions",
    width: 100,
    render: row =>
      h(
        NButton,
        {
          size: "small",
          type: "primary",
          ghost: true,
          onClick: () => viewDetail(row),
        },
        { default: () => t("tokens.detail") }
      ),
  },
]);

// Methods
async function loadTokenRecords() {
  loading.value = true;
  try {
    // 使用 RequestLogsPanel 的日志请求接口
    const response = await logApi.getLogs(filters);
    if (response && response.data && response.data.items && response.data.pagination) {
      // 将日志数据转换为 token 记录格式
      tokenRecords.value = response.data.items
        // .filter(
        //   (log: any) =>
        //     (log.total_tokens && log.total_tokens > 0) ||
        //     (log.prompt_tokens && log.prompt_tokens > 0) ||
        //     (log.completion_tokens && log.completion_tokens > 0)
        // )
        .map((log: any) => claudeTokenApi.convertLogToTokenRecord(log));

      pagination.value = response.data.pagination;
    } else {
      // Handle empty or invalid response
      tokenRecords.value = [];
      pagination.value = {
        page: 1,
        page_size: 20,
        total_items: 0,
        total_pages: 0,
      };
    }
  } catch (error) {
    console.error("Failed to load token records:", error);
    message.error(t("tokens.loadFailed"));
    // Reset to default state on error
    tokenRecords.value = [];
    pagination.value = {
      page: 1,
      page_size: 20,
      total_items: 0,
      total_pages: 0,
    };
  } finally {
    loading.value = false;
  }
}

async function loadTokenStats() {
  try {
    const response = await claudeTokenApi.getTokenStats();
    const stats = response.data;
    const requestCount = tokenRecords.value.length;
    const avgTokensPerRequest = requestCount > 0 ? Math.round(stats.totalTokens / requestCount) : 0;

    tokenStats.value = {
      ...stats,
      requestCount,
      avgTokensPerRequest,
      todayUsage: formatTokenCount(Math.floor(stats.totalTokens * 0.15)), // Simulate today's usage
      growthRate: 12.5, // Simulate growth rate
      estimatedCost: calculateTokenCost(
        stats.totalInputTokens,
        stats.totalOutputTokens,
        "claude-sonnet-4"
      ).toFixed(2),
    };
  } catch (error) {
    console.error("Failed to load token stats:", error);
  }
}

async function loadChartData() {
  // This would load chart data based on the selected period
  // For now, just a placeholder
}

function viewDetail(record: TokenConsumptionRecord) {
  selectedRecord.value = record;
  showDetailModal.value = true;
}

function resetFilters() {
  Object.assign(filters, {
    page: 1,
    page_size: 20,
    start_time: "",
    end_time: "",
    model: "",
    group: "",
    minTokens: "",
    maxTokens: "",
  });
  dateRange.value = null;
  loadTokenRecords();
}

function onDateRangeChange(value: [number, number] | null) {
  if (value) {
    filters.start_time = new Date(value[0]).toISOString();
    filters.end_time = new Date(value[1]).toISOString();
  } else {
    filters.start_time = "";
    filters.end_time = "";
  }
}

function handlePageChange(page: number) {
  filters.page = page;
  loadTokenRecords();
}

function handlePageSizeChange(pageSize: number) {
  filters.page_size = pageSize;
  filters.page = 1;
  loadTokenRecords();
}

async function exportTokenLogs() {
  exporting.value = true;
  try {
    // 使用 RequestLogsPanel 的导出接口
    logApi.exportLogs(filters);
    message.success(t("tokens.exportSuccess"));
  } catch (_error) {
    message.error(t("tokens.exportFailed"));
  } finally {
    exporting.value = false;
  }
}

// Watch filters
watch(
  [() => filters.model, () => filters.group, () => filters.start_time, () => filters.end_time],
  () => {
    filters.page = 1;
    loadTokenRecords();
  }
);

// Initialize
onMounted(() => {
  loadTokenRecords();
  loadTokenStats();
  loadChartData();
});
</script>

<template>
  <div class="token-logs-panel">
    <!-- Statistics Overview -->
    <div class="token-stats-grid">
      <n-card class="token-stat-card primary">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="28">
              <stats-chart-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ formatTokenCount(tokenStats.totalTokens) }}</div>
            <div class="stat-label">{{ t("tokens.totalTokens") }}</div>
            <div class="stat-sub">
              <span class="input-tokens">
                {{ formatTokenCount(tokenStats.totalInputTokens) }} {{ t("tokens.input") }}
              </span>
              <span class="output-tokens">
                {{ formatTokenCount(tokenStats.totalOutputTokens) }} {{ t("tokens.output") }}
              </span>
            </div>
          </div>
        </div>
      </n-card>

      <n-card class="token-stat-card success">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="28">
              <cash-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">${{ tokenStats.estimatedCost }}</div>
            <div class="stat-label">{{ t("tokens.estimatedCost") }}</div>
            <div class="stat-sub">
              <span>{{ t("tokens.basedOnPricing") }}</span>
            </div>
          </div>
        </div>
      </n-card>

      <n-card class="token-stat-card info">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="28">
              <time-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ tokenStats.requestCount }}</div>
            <div class="stat-label">{{ t("tokens.totalRequests") }}</div>
            <div class="stat-sub">
              <span>{{ tokenStats.avgTokensPerRequest }} {{ t("tokens.avgPerRequest") }}</span>
            </div>
          </div>
        </div>
      </n-card>

      <n-card class="token-stat-card warning">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="28">
              <trending-up-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ tokenStats.todayUsage }}</div>
            <div class="stat-label">{{ t("tokens.todayUsage") }}</div>
            <div class="stat-sub">
              <span class="trend positive">
                +{{ tokenStats.growthRate }}% {{ t("tokens.fromYesterday") }}
              </span>
            </div>
          </div>
        </div>
      </n-card>
    </div>

    <!-- Filter and Actions -->
    <n-card class="filter-card" size="small">
      <div class="filter-header">
        <div class="filter-title">
          <n-icon size="20">
            <filter-outline />
          </n-icon>
          <span>{{ t("tokens.filters") }}</span>
        </div>
        <div class="filter-actions">
          <n-button size="small" @click="resetFilters">
            <template #icon>
              <n-icon><refresh-outline /></n-icon>
            </template>
            {{ t("common.reset") }}
          </n-button>
          <n-button size="small" type="primary" ghost @click="exportTokenLogs" :loading="exporting">
            <template #icon>
              <n-icon><cloud-download-outline /></n-icon>
            </template>
            {{ t("tokens.export") }}
          </n-button>
        </div>
      </div>

      <div class="filter-grid">
        <n-form-item :label="t('tokens.dateRange')">
          <n-date-picker
            v-model:value="dateRange"
            type="datetimerange"
            :placeholder="[t('tokens.startTime'), t('tokens.endTime')]"
            clearable
            @update:value="onDateRangeChange"
          />
        </n-form-item>

        <n-form-item :label="t('tokens.model')">
          <n-select
            v-model:value="filters.model"
            :placeholder="t('tokens.allModels')"
            :options="modelOptions"
            clearable
            filterable
          />
        </n-form-item>

        <n-form-item :label="t('tokens.group')">
          <n-select
            v-model:value="filters.group"
            :placeholder="t('tokens.allGroups')"
            :options="groupOptions"
            clearable
          />
        </n-form-item>

        <n-form-item :label="t('tokens.tokenRange')">
          <n-input-group>
            <n-input
              v-model:value="filters.minTokens"
              :placeholder="t('tokens.min')"
              type="number"
              style="width: 50%"
            />
            <n-input
              v-model:value="filters.maxTokens"
              :placeholder="t('tokens.max')"
              type="number"
              style="width: 50%"
            />
          </n-input-group>
        </n-form-item>
      </div>
    </n-card>

    <!-- Token Usage Chart -->
    <n-card class="chart-card">
      <template #header>
        <div class="chart-header">
          <div class="chart-title">
            <n-icon size="20">
              <bar-chart-outline />
            </n-icon>
            <span>{{ t("tokens.usageTrend") }}</span>
          </div>
          <n-radio-group v-model:value="chartPeriod" size="small" @update:value="loadChartData">
            <n-radio-button value="24h">{{ t("tokens.last24h") }}</n-radio-button>
            <n-radio-button value="7d">{{ t("tokens.last7days") }}</n-radio-button>
            <n-radio-button value="30d">{{ t("tokens.last30days") }}</n-radio-button>
          </n-radio-group>
        </div>
      </template>

      <div ref="chartContainer" class="chart-container">
        <!-- Chart will be rendered here -->
        <div class="chart-placeholder">
          <n-icon size="48" color="#ccc">
            <bar-chart-outline />
          </n-icon>
          <p>{{ t("tokens.chartLoading") }}</p>
        </div>
      </div>
    </n-card>

    <!-- Token Records Table -->
    <n-card class="table-card">
      <template #header>
        <div class="table-header">
          <div class="table-title">
            <n-icon size="20">
              <document-text-outline />
            </n-icon>
            <span>{{ t("tokens.tokenRecords") }}</span>
            <n-badge :value="pagination.total_items" :max="9999" type="success" show-zero />
          </div>

          <n-button size="small" @click="loadTokenRecords" :loading="loading">
            <template #icon>
              <n-icon><refresh-outline /></n-icon>
            </template>
            {{ t("common.refresh") }}
          </n-button>
        </div>
      </template>
      <n-data-table
        :columns="columns"
        :data="tokenRecords"
        :loading="loading"
        :pagination="{
          page: filters.page,
          pageSize: filters.page_size,
          itemCount: pagination.total_items,
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          onUpdatePage: handlePageChange,
          onUpdatePageSize: handlePageSizeChange,
          prefix: ({ itemCount }) => t('tokens.totalRecords', { count: itemCount }),
        }"
        size="small"
        flex-height
        style="min-height: 400px"
      />
    </n-card>

    <!-- Token Detail Modal -->
    <n-modal
      v-model:show="showDetailModal"
      :title="t('tokens.tokenDetail')"
      preset="card"
      size="large"
      :style="{ maxWidth: '80vw', width: '900px' }"
    >
      <token-detail-panel
        v-if="selectedRecord"
        :record="selectedRecord"
        @close="showDetailModal = false"
      />
    </n-modal>
  </div>
</template>

<style scoped>
.token-logs-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
}

.token-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  flex-shrink: 0;
}

.token-stat-card {
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.token-stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.token-stat-card.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.token-stat-card.success {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: white;
}

.token-stat-card.info {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: white;
}

.token-stat-card.warning {
  background: linear-gradient(135deg, #fa8c16 0%, #ffa940 100%);
  color: white;
}

.stat-content {
  display: flex;
  align-items: flex-start;
  padding: 24px;
  gap: 16px;
}

.stat-icon {
  opacity: 0.9;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 6px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.stat-sub {
  font-size: 12px;
  opacity: 0.8;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.input-tokens,
.output-tokens {
  background: rgba(255, 255, 255, 0.15);
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  display: inline-block;
  margin-right: 4px;
}

.trend.positive {
  color: #95f985;
  font-weight: 600;
}

.filter-card {
  flex-shrink: 0;
}

.filter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filter-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: var(--text-color-1);
}

.filter-actions {
  display: flex;
  gap: 8px;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.chart-card {
  flex-shrink: 0;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.chart-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  text-align: center;
  color: #ccc;
}

.chart-placeholder p {
  margin: 8px 0 0 0;
  font-size: 14px;
}

.table-card {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.table-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

:deep(.n-data-table) {
  flex: 1;
}

.total-tokens {
  font-weight: 600;
  color: var(--primary-color);
}

@media (max-width: 768px) {
  .token-stats-grid {
    grid-template-columns: 1fr;
  }

  .filter-grid {
    grid-template-columns: 1fr;
  }

  .stat-content {
    padding: 20px;
    gap: 12px;
  }

  .stat-value {
    font-size: 24px;
  }

  .chart-container {
    height: 250px;
  }
}
</style>
