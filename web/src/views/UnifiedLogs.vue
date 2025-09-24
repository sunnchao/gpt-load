<script setup lang="ts">
import { formatTokenCount } from "@/api/claude-tokens";
import RequestLogsPanel from "@/components/logs/RequestLogsPanel.vue";
import TokenLogsPanel from "@/components/logs/TokenLogsPanel.vue";
import { AnalyticsOutline, DocumentTextOutline, StatsChartOutline } from "@vicons/ionicons5";
import { NBadge, NIcon, NTabPane, NTabs } from "naive-ui";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

// Current active tab
const activeTab = ref<"requests" | "tokens">("requests");

// Stats data
const requestStats = ref({
  total: 0,
  success: 0,
  failed: 0,
});

const tokenStats = ref({
  totalTokens: 0,
  totalInputTokens: 0,
  totalOutputTokens: 0,
});

// Load statistics on mount
onMounted(() => {
  loadStats();
});

async function loadStats() {
  // This will be implemented to load basic stats for badges
  // For now, using placeholder data
}
</script>

<template>
  <div class="unified-logs-page">
    <!-- Header with tabs and controls -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <n-icon size="24" class="title-icon">
            <document-text-outline />
          </n-icon>
          {{ t("logs.title") }}
        </h1>
        <p class="page-subtitle">{{ t("logs.subtitle") }}</p>
      </div>

      <!-- Tab navigation -->
      <n-tabs v-model:value="activeTab" type="segment" class="logs-tabs" animated>
        <n-tab-pane name="requests" :tab="t('logs.requestLogs')">
          <template #tab>
            <div class="tab-content">
              <n-icon size="18">
                <analytics-outline />
              </n-icon>
              <span>{{ t("logs.requestLogs") }}</span>
              <n-badge
                v-if="requestStats.total > 0"
                :value="requestStats.total"
                :max="999"
                type="info"
                show-zero
              />
            </div>
          </template>
          <div class="tab-content-wrapper">
            <request-logs-panel />
          </div>
        </n-tab-pane>

        <n-tab-pane name="tokens" :tab="t('logs.tokenUsage')">
          <template #tab>
            <div class="tab-content">
              <n-icon size="18">
                <stats-chart-outline />
              </n-icon>
              <span>{{ t("logs.tokenUsage") }}</span>
              <n-badge
                v-if="tokenStats.totalTokens > 0"
                :value="formatTokenCount(tokenStats.totalTokens)"
                type="success"
              />
            </div>
          </template>
          <div class="tab-content-wrapper">
            <token-logs-panel />
          </div>
        </n-tab-pane>
      </n-tabs>
    </div>

    <!-- Content area is now managed by n-tabs -->
  </div>
</template>

<style scoped>
.unified-logs-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-color);
}

.page-header {
  padding: 24px 32px 0;
  background: var(--card-color);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.header-content {
  margin-bottom: 24px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 28px;
  font-weight: 700;
  color: var(--text-color-1);
  margin: 0 0 8px 0;
}

.title-icon {
  color: var(--primary-color);
}

.page-subtitle {
  color: var(--text-color-3);
  font-size: 14px;
  margin: 0;
}

.logs-tabs {
  --n-tab-font-size: 16px;
  --n-tab-font-weight: 600;
  --n-tab-padding: 16px 24px;
  --n-tab-gap: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tab-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-content-wrapper {
  flex: 1;
  padding: 24px 32px;
  overflow: hidden;
}

/* Dark mode support */
.dark .unified-logs-page {
  --bg-color: #1a1d23;
  --card-color: #0f1115;
  --border-color: rgba(255, 255, 255, 0.08);
  --text-color-1: #e8e8e8;
  --text-color-3: #888888;
  --primary-color: #667eea;
}

/* Light mode support */
.unified-logs-page {
  --bg-color: #f5f5f7;
  --card-color: #ffffff;
  --border-color: rgba(0, 0, 0, 0.08);
  --text-color-1: #1a1a1a;
  --text-color-3: #8a8a8a;
  --primary-color: #667eea;
}

/* Responsive design */
@media (max-width: 768px) {
  .page-header {
    padding: 16px 20px 0;
  }

  .tab-content-wrapper {
    padding: 16px 20px;
  }

  .page-title {
    font-size: 24px;
  }

  .logs-tabs {
    --n-tab-font-size: 14px;
    --n-tab-padding: 12px 16px;
  }

  .tab-content {
    gap: 6px;
  }
}
</style>
