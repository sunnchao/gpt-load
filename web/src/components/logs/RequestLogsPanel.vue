<script setup lang="ts">
import { logApi } from "@/api/logs";
import RequestDetailPanel from "@/components/logs/RequestDetailPanel.vue";
import type { LogFilter, Pagination, RequestLog } from "@/types/models";
import {
  AnalyticsOutline,
  CheckmarkDoneOutline,
  CloseCircleOutline,
  DownloadOutline,
  ListOutline,
  RefreshOutline,
  Search,
  TimeOutline,
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
  NModal,
  NSelect,
  useMessage,
  type DataTableColumns,
} from "naive-ui";
import { computed, h, onMounted, reactive, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const message = useMessage();

// Data
const loading = ref(false);
const exporting = ref(false);
const logs = ref<RequestLog[]>([]);
const selectedLog = ref<RequestLog | null>(null);
const showDetailModal = ref(false);

// Filters and pagination
const filters = reactive<LogFilter>({
  page: 1,
  page_size: 20,
  group_name: null,
  is_success: null,
  model: "",
  start_time: null,
  end_time: null,
});

const pagination = ref<Pagination>({
  page: 1,
  page_size: 20,
  total_items: 0,
  total_pages: 0,
});

const dateRange = ref<[number, number] | null>(null);

// Statistics
const stats = ref({
  successCount: 0,
  errorCount: 0,
  avgDuration: 0,
  totalRequests: 0,
  successTrend: 0,
  errorTrend: 0,
  durationTrend: 0,
  totalTrend: 0,
});

// Options
const groupOptions = ref([
  { label: "OpenAI", value: "openai" },
  { label: "Claude", value: "claude" },
  { label: "Gemini", value: "gemini" },
]);

const statusOptions = [
  { label: t("logs.successful"), value: true },
  { label: t("logs.failed"), value: false },
];

// Table columns
const columns = computed<DataTableColumns<RequestLog>>(() => [
  {
    title: t("logs.timestamp"),
    key: "timestamp",
    width: 160,
    render: row => new Date(row.timestamp).toLocaleString(),
  },
  {
    title: t("logs.status"),
    key: "is_success",
    width: 80,
    render: row =>
      h(
        "div",
        { class: `status-badge ${row.is_success ? "success" : "error"}` },
        row.is_success ? t("logs.success") : t("logs.failed")
      ),
  },
  {
    title: t("logs.group"),
    key: "group_name",
    width: 120,
    ellipsis: { tooltip: true },
  },
  {
    title: t("logs.model"),
    key: "model",
    width: 140,
    ellipsis: { tooltip: true },
  },
  {
    title: t("logs.duration"),
    key: "duration_ms",
    width: 100,
    render: row => `${row.duration_ms}ms`,
  },
  {
    title: t("logs.tokens"),
    key: "total_tokens",
    width: 100,
    render: row => row.total_tokens || "-",
  },
  {
    title: t("logs.actions"),
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
        { default: () => t("logs.detail") }
      ),
  },
]);

// Methods
async function loadLogs() {
  loading.value = true;
  try {
    const response = await logApi.getLogs(filters);
    if (response && response.data.items && response.data.pagination) {
      logs.value = response.data.items;
      pagination.value = response.data.pagination;
      await loadStats();
    } else {
      // Handle empty or invalid response
      logs.value = [];
      pagination.value = {
        page: 1,
        page_size: 20,
        total_items: 0,
        total_pages: 0,
      };
    }
  } catch (error) {
    console.error("Failed to load logs:", error);
    message.error(t("logs.loadFailed"));
    // Reset to default state on error
    logs.value = [];
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

async function loadStats() {
  try {
    // Load basic statistics
    const successCount = logs.value.filter(log => log.is_success).length;
    const errorCount = logs.value.length - successCount;
    const avgDuration =
      logs.value.length > 0
        ? Math.round(logs.value.reduce((sum, log) => sum + log.duration_ms, 0) / logs.value.length)
        : 0;

    stats.value = {
      successCount,
      errorCount,
      avgDuration,
      totalRequests: logs.value.length,
      successTrend: 5.2,
      errorTrend: -2.1,
      durationTrend: -8.3,
      totalTrend: 12.7,
    };
  } catch (error) {
    console.error("Failed to load stats:", error);
  }
}

function viewDetail(log: RequestLog) {
  selectedLog.value = log;
  showDetailModal.value = true;
}

function resetFilters() {
  Object.assign(filters, {
    page: 1,
    page_size: 20,
    group_name: null,
    is_success: null,
    model: "",
    start_time: null,
    end_time: null,
  });
  dateRange.value = null;
  loadLogs();
}

function onDateRangeChange(value: [number, number] | null) {
  if (value) {
    filters.start_time = new Date(value[0]).toISOString();
    filters.end_time = new Date(value[1]).toISOString();
  } else {
    filters.start_time = null;
    filters.end_time = null;
  }
}

function handlePageChange(page: number) {
  filters.page = page;
  loadLogs();
}

function handlePageSizeChange(pageSize: number) {
  filters.page_size = pageSize;
  filters.page = 1;
  loadLogs();
}

function getRowClassName(row: RequestLog) {
  return row.is_success ? "" : "error-row";
}

async function exportLogs() {
  exporting.value = true;
  try {
    logApi.exportLogs(filters);
    message.success(t("logs.exportSuccess"));
  } catch (error) {
    message.error(t("logs.exportFailed"));
  } finally {
    exporting.value = false;
  }
}

// Watch filters
watch(
  [() => filters.group_name, () => filters.is_success, () => filters.model],
  () => {
    filters.page = 1;
    loadLogs();
  },
  { deep: true }
);

// Initialize
onMounted(() => {
  loadLogs();
});
</script>

<template>
  <div class="request-logs-panel">
    <!-- Filter controls -->
    <n-card class="filter-card" size="small">
      <div class="filter-header">
        <div class="filter-title">
          <n-icon size="20">
            <search />
          </n-icon>
          <span>{{ t("logs.filters") }}</span>
        </div>
        <n-button size="small" type="primary" ghost @click="resetFilters">
          <template #icon>
            <n-icon><refresh-outline /></n-icon>
          </template>
          {{ t("common.reset") }}
        </n-button>
      </div>

      <div class="filter-grid">
        <n-form-item :label="t('logs.groupName')">
          <n-select
            v-model:value="filters.group_name"
            :placeholder="t('logs.allGroups')"
            :options="groupOptions"
            clearable
            filterable
          />
        </n-form-item>

        <n-form-item :label="t('logs.status')">
          <n-select
            v-model:value="filters.is_success"
            :placeholder="t('logs.allStatuses')"
            :options="statusOptions"
            clearable
          />
        </n-form-item>

        <n-form-item :label="t('logs.model')">
          <n-input
            v-model:value="filters.model"
            :placeholder="t('logs.modelPlaceholder')"
            clearable
          />
        </n-form-item>

        <n-form-item :label="t('logs.dateRange')">
          <n-date-picker
            v-model:value="dateRange"
            type="datetimerange"
            :placeholder="[t('logs.startTime'), t('logs.endTime')]"
            clearable
            @update:value="onDateRangeChange"
          />
        </n-form-item>
      </div>
    </n-card>

    <!-- Statistics cards -->
    <div class="stats-grid">
      <n-card class="stat-card success">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="24">
              <checkmark-done-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.successCount }}</div>
            <div class="stat-label">{{ t("logs.successfulRequests") }}</div>
          </div>
          <div class="stat-trend" :class="{ positive: stats.successTrend > 0 }">
            {{ stats.successTrend > 0 ? "+" : "" }}{{ stats.successTrend }}%
          </div>
        </div>
      </n-card>

      <n-card class="stat-card error">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="24">
              <close-circle-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.errorCount }}</div>
            <div class="stat-label">{{ t("logs.failedRequests") }}</div>
          </div>
          <div class="stat-trend" :class="{ negative: stats.errorTrend < 0 }">
            {{ stats.errorTrend > 0 ? "+" : "" }}{{ stats.errorTrend }}%
          </div>
        </div>
      </n-card>

      <n-card class="stat-card info">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="24">
              <time-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.avgDuration }}ms</div>
            <div class="stat-label">{{ t("logs.averageResponse") }}</div>
          </div>
          <div class="stat-trend" :class="{ positive: stats.durationTrend < 0 }">
            {{ stats.durationTrend > 0 ? "+" : "" }}{{ stats.durationTrend }}%
          </div>
        </div>
      </n-card>

      <n-card class="stat-card primary">
        <div class="stat-content">
          <div class="stat-icon">
            <n-icon size="24">
              <analytics-outline />
            </n-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.totalRequests }}</div>
            <div class="stat-label">{{ t("logs.totalRequests") }}</div>
          </div>
          <div class="stat-trend" :class="{ positive: stats.totalTrend > 0 }">
            {{ stats.totalTrend > 0 ? "+" : "" }}{{ stats.totalTrend }}%
          </div>
        </div>
      </n-card>
    </div>

    <!-- Data table -->
    <n-card class="table-card">
      <template #header>
        <div class="table-header">
          <div class="table-title">
            <n-icon size="20">
              <list-outline />
            </n-icon>
            <span>{{ t("logs.requestHistory") }}</span>
            <n-badge :value="pagination?.total_items || 0" :max="9999" type="info" show-zero />
          </div>

          <div class="table-actions">
            <n-button size="small" @click="loadLogs" :loading="loading">
              <template #icon>
                <n-icon><refresh-outline /></n-icon>
              </template>
              {{ t("common.refresh") }}
            </n-button>

            <n-button size="small" type="primary" ghost @click="exportLogs" :loading="exporting">
              <template #icon>
                <n-icon><download-outline /></n-icon>
              </template>
              {{ t("logs.export") }}
            </n-button>
          </div>
        </div>
      </template>

      <n-data-table
        :columns="columns"
        :data="logs"
        :loading="loading"
        :pagination="{
          page: filters.page,
          pageSize: filters.page_size,
          itemCount: pagination?.total_items || 0,
          showSizePicker: true,
          pageSizes: [10, 20, 50, 100],
          onUpdatePage: handlePageChange,
          onUpdatePageSize: handlePageSizeChange,
          prefix: ({ itemCount }) => t('logs.totalItems', { count: itemCount }),
        }"
        :row-class-name="getRowClassName"
        size="small"
        flex-height
        style="min-height: 400px"
      />
    </n-card>

    <!-- Request detail modal -->
    <n-modal
      v-model:show="showDetailModal"
      :title="t('logs.requestDetail')"
      preset="card"
      size="huge"
      :style="{ maxWidth: '90vw', width: '1200px' }"
    >
      <request-detail-panel
        v-if="selectedLog"
        :log="selectedLog"
        @close="showDetailModal = false"
      />
    </n-modal>
  </div>
</template>

<style scoped>
.request-logs-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
  height: 100%;
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

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  flex-shrink: 0;
}

.stat-card {
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.stat-card.success {
  background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  color: white;
}

.stat-card.error {
  background: linear-gradient(135deg, #ff4d4f 0%, #ff7875 100%);
  color: white;
}

.stat-card.info {
  background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  color: white;
}

.stat-card.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 20px;
  gap: 16px;
}

.stat-icon {
  opacity: 0.8;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.stat-trend {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.2);
}

.stat-trend.positive {
  background: rgba(255, 255, 255, 0.2);
}

.stat-trend.negative {
  background: rgba(255, 255, 255, 0.2);
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

.table-actions {
  display: flex;
  gap: 8px;
}

:deep(.n-data-table) {
  flex: 1;
}

:deep(.error-row) {
  background: rgba(255, 77, 79, 0.05);
}

.status-badge {
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  text-align: center;
}

.status-badge.success {
  background: #f6ffed;
  color: #52c41a;
  border: 1px solid #b7eb8f;
}

.status-badge.error {
  background: #fff2f0;
  color: #ff4d4f;
  border: 1px solid #ffb3b0;
}

.dark .status-badge.success {
  background: rgba(82, 196, 26, 0.1);
  color: #95de64;
  border: 1px solid rgba(82, 196, 26, 0.3);
}

.dark .status-badge.error {
  background: rgba(255, 77, 79, 0.1);
  color: #ff7875;
  border: 1px solid rgba(255, 77, 79, 0.3);
}

@media (max-width: 768px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-content {
    padding: 16px;
    gap: 12px;
  }

  .stat-value {
    font-size: 20px;
  }
}
</style>
