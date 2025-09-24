<template>
  <div class="request-detail-panel">
    <n-card title="请求详情" :bordered="false">
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
              <n-text code>{{ record.id }}</n-text>
            </n-descriptions-item>
            <n-descriptions-item label="分组">
              <n-tag v-if="record.group_name">{{ record.group_name }}</n-tag>
              <n-text depth="3" v-else>-</n-text>
            </n-descriptions-item>
            <n-descriptions-item label="模型">
              <n-tag type="info">{{ record.model || '-' }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="record.is_success ? 'success' : 'error'">
                {{ record.is_success ? '成功' : '失败' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="状态码">
              <n-tag :type="getStatusCodeType(record.status_code)">
                {{ record.status_code }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="耗时">
              {{ record.duration_ms }}ms
            </n-descriptions-item>
            <n-descriptions-item label="请求类型">
              <n-tag :type="record.request_type === 'final' ? 'success' : 'warning'">
                {{ record.request_type === 'final' ? '最终请求' : '重试请求' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="是否流式">
              <n-tag :type="record.is_stream ? 'info' : 'default'">
                {{ record.is_stream ? '流式' : '非流式' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="来源IP">
              <n-text code>{{ record.source_ip || '-' }}</n-text>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <!-- 请求路径和上游地址 -->
        <n-card title="网络信息" size="small" embedded>
          <n-descriptions :columns="1" bordered>
            <n-descriptions-item label="请求路径">
              <n-text code style="word-break: break-all;">
                {{ record.request_path }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item v-if="record.upstream_addr" label="上游地址">
              <n-text code style="word-break: break-all;">
                {{ record.upstream_addr }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item v-if="record.user_agent" label="User Agent">
              <n-text depth="3" style="word-break: break-all;">
                {{ record.user_agent }}
              </n-text>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>

        <!-- Token 统计 -->
        <n-card v-if="hasTokenData" title="Token 统计" size="small" embedded>
          <n-grid :cols="3" :x-gap="16">
            <n-grid-item v-if="record.prompt_tokens">
              <n-statistic label="Prompt Token" :value="record.prompt_tokens">
                <template #suffix>
                  <n-icon :component="DocumentTextOutline" />
                </template>
              </n-statistic>
            </n-grid-item>
            <n-grid-item v-if="record.completion_tokens">
              <n-statistic label="Completion Token" :value="record.completion_tokens">
                <template #suffix>
                  <n-icon :component="CreateOutline" />
                </template>
              </n-statistic>
            </n-grid-item>
            <n-grid-item v-if="record.total_tokens">
              <n-statistic label="总计 Token" :value="record.total_tokens">
                <template #suffix>
                  <n-icon :component="BarChartOutline" />
                </template>
              </n-statistic>
            </n-grid-item>
          </n-grid>

          <!-- 扩展 Token 信息 -->
          <template v-if="hasExtendedTokenData">
            <n-divider />
            <n-descriptions :columns="2" bordered size="small">
              <n-descriptions-item
                v-if="record.cached_prompt_tokens"
                label="缓存 Prompt Token"
              >
                {{ record.cached_prompt_tokens }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="record.cached_completion_tokens"
                label="缓存 Completion Token"
              >
                {{ record.cached_completion_tokens }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="record.reasoning_tokens"
                label="推理 Token"
              >
                {{ record.reasoning_tokens }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="record.audio_tokens"
                label="音频 Token"
              >
                {{ record.audio_tokens }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="record.image_tokens"
                label="图像 Token"
              >
                {{ record.image_tokens }}
              </n-descriptions-item>
            </n-descriptions>
          </template>
        </n-card>

        <!-- 错误信息 -->
        <n-card v-if="record.error_message" title="错误信息" size="small" embedded>
          <n-alert type="error" :show-icon="false">
            {{ record.error_message }}
          </n-alert>
        </n-card>

        <!-- 请求体 -->
        <n-card v-if="record.request_body" title="请求体" size="small" embedded>
          <n-code
            :code="record.request_body"
            language="json"
            show-line-numbers
            style="max-height: 300px; overflow-y: auto;"
          />
        </n-card>

        <!-- 响应体 -->
        <n-card v-if="record.response_body" title="响应体" size="small" embedded>
          <n-code
            :code="record.response_body"
            language="json"
            show-line-numbers
            style="max-height: 300px; overflow-y: auto;"
          />
        </n-card>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import {
  NCard,
  NButton,
  NIcon,
  NSpace,
  NDescriptions,
  NDescriptionsItem,
  NText,
  NTag,
  NStatistic,
  NGrid,
  NGridItem,
  NDivider,
  NAlert,
  NCode,
} from 'naive-ui';
import {
  CloseOutline,
  DocumentTextOutline,
  CreateOutline,
  BarChartOutline,
} from '@vicons/ionicons5';
import type { RequestLog } from '@/types/models';

interface Props {
  record: RequestLog;
}

const props = defineProps<Props>();
defineEmits<{
  close: [];
}>();

// 格式化日期时间
const formatDateTime = (timestamp: string) => {
  const date = new Date(timestamp);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

// 获取状态码标签类型
const getStatusCodeType = (statusCode: number) => {
  if (statusCode >= 200 && statusCode < 300) return 'success';
  if (statusCode >= 300 && statusCode < 400) return 'info';
  if (statusCode >= 400 && statusCode < 500) return 'warning';
  return 'error';
};

// 检查是否有 token 数据
const hasTokenData = computed(() => {
  return !!(props.record.prompt_tokens || props.record.completion_tokens || props.record.total_tokens);
});

// 检查是否有扩展的 token 数据
const hasExtendedTokenData = computed(() => {
  return !!(
    props.record.cached_prompt_tokens ||
    props.record.cached_completion_tokens ||
    props.record.reasoning_tokens ||
    props.record.audio_tokens ||
    props.record.image_tokens
  );
});
</script>

<style scoped>
.request-detail-panel {
  width: 100%;
  max-width: 900px;
}

.n-statistic {
  text-align: center;
}

.n-descriptions-item {
  --n-th-padding: 8px 12px;
  --n-td-padding: 8px 12px;
}
</style>