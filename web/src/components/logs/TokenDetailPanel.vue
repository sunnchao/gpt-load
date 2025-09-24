<script setup lang="ts">
import { type TokenConsumptionRecord } from "@/api/claude-tokens";
import { calculateTokenCost, formatTokenCount } from "@/utils/token-cost";
import {
  BarChartOutline,
  CashOutline,
  CloseOutline,
  CreateOutline,
  DocumentTextOutline,
} from "@vicons/ionicons5";
import {
  NAlert,
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NGrid,
  NGridItem,
  NIcon,
  NSpace,
  NStatistic,
  NTag,
  NText,
} from "naive-ui";
import { computed } from "vue";

interface Props {
  record: TokenConsumptionRecord;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
}>();

// 格式化日期时间
const formatDateTime = (timestamp: string) => {
  const date = new Date(timestamp);
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

// 计算预估成本
const estimatedCost = computed(() => {
  return calculateTokenCost(
    props.record.inputTokens || 0,
    props.record.outputTokens || 0,
    props.record.model || "",
    props.record.cacheCreationTokens,
    props.record.cacheReadTokens
  );
});

// 获取服务等级标签类型
const getServiceTierType = (tier: string) => {
  switch (tier?.toLowerCase()) {
    case "scale":
      return "success";
    case "premium":
      return "warning";
    default:
      return "default";
  }
};
</script>

<template>
  <div class="token-detail-panel">
    <n-card title="Token 使用详情" :bordered="false">
      <template #header-extra>
        <n-button quaternary circle @click="$emit('close')">
          <template #icon>
            <n-icon :component="CloseOutline" />
          </template>
        </n-button>
      </template>

      <n-space vertical :size="24">
        <!-- 基础信息 -->
        <n-card title="基础信息" size="small" embedded>
          <n-descriptions :columns="2" bordered>
            <n-descriptions-item label="请求时间">
              {{ formatDateTime(record.timestamp) }}
            </n-descriptions-item>
            <n-descriptions-item label="请求ID">
              <n-text code>{{ record.requestId }}</n-text>
            </n-descriptions-item>
            <n-descriptions-item label="模型">
              <n-tag type="info">{{ record.model }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="分组">
              <n-tag v-if="record.groupName">{{ record.groupName }}</n-tag>
              <n-text depth="3" v-else>-</n-text>
            </n-descriptions-item>
            <n-descriptions-item label="请求耗时">
              {{ record.duration ? `${record.duration}ms` : "-" }}
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="record.isSuccess ? 'success' : 'error'">
                {{ record.isSuccess ? "成功" : "失败" }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <!-- Token 使用详情 -->
        <n-card title="Token 使用统计" size="small" embedded>
          <n-space vertical :size="16">
            <!-- 主要 Token 统计 -->
            <n-grid :cols="3" :x-gap="16">
              <n-grid-item>
                <n-statistic label="输入 Token" :value="record.inputTokens">
                  <template #suffix>
                    <n-icon :component="DocumentTextOutline" />
                  </template>
                </n-statistic>
              </n-grid-item>
              <n-grid-item>
                <n-statistic label="输出 Token" :value="record.outputTokens">
                  <template #suffix>
                    <n-icon :component="CreateOutline" />
                  </template>
                </n-statistic>
              </n-grid-item>
              <n-grid-item>
                <n-statistic label="总计 Token" :value="record.totalTokens">
                  <template #suffix>
                    <n-icon :component="BarChartOutline" />
                  </template>
                </n-statistic>
              </n-grid-item>
            </n-grid>

            <!-- 扩展 Token 信息 -->
            <n-divider />
            <n-descriptions :columns="2" bordered size="small">
              <n-descriptions-item v-if="record.cachedPromptTokens" label="缓存 Prompt Token">
                {{ formatTokenCount(record.cachedPromptTokens) }}
              </n-descriptions-item>
              <n-descriptions-item v-if="record.reasoningTokens" label="推理 Token">
                {{ formatTokenCount(record.reasoningTokens) }}
              </n-descriptions-item>
              <n-descriptions-item v-if="record.audioTokens" label="音频 Token">
                {{ formatTokenCount(record.audioTokens) }}
              </n-descriptions-item>
              <n-descriptions-item v-if="record.imageTokens" label="图像 Token">
                {{ formatTokenCount(record.imageTokens) }}
              </n-descriptions-item>
              <n-descriptions-item v-if="record.cacheCreationTokens" label="缓存创建 Token">
                {{ formatTokenCount(record.cacheCreationTokens) }}
              </n-descriptions-item>
              <n-descriptions-item v-if="record.cacheReadTokens" label="缓存读取 Token">
                {{ formatTokenCount(record.cacheReadTokens) }}
              </n-descriptions-item>
            </n-descriptions>
          </n-space>
        </n-card>

        <!-- 成本估算 -->
        <n-card title="成本估算" size="small" embedded>
          <n-alert type="info" :show-icon="false">
            <template #icon>
              <n-icon :component="CashOutline" />
            </template>
            <n-space>
              <n-text>
                预估成本:
                <n-text strong type="success">${{ estimatedCost.toFixed(6) }}</n-text>
              </n-text>
              <n-text depth="3">(基于 {{ record.model }} 定价)</n-text>
            </n-space>
          </n-alert>
        </n-card>

        <!-- 服务等级 -->
        <n-card v-if="record.serviceTier" title="服务信息" size="small" embedded>
          <n-descriptions :columns="1" bordered size="small">
            <n-descriptions-item label="服务等级">
              <n-tag :type="getServiceTierType(record.serviceTier)">
                {{ record.serviceTier }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.token-detail-panel {
  width: 100%;
  max-width: 800px;
}

.n-statistic {
  text-align: center;
}

.n-descriptions-item {
  --n-th-padding: 8px 12px;
  --n-td-padding: 8px 12px;
}
</style>
