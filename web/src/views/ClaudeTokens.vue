<script setup lang="ts">
import ClaudeTokenTracker from "@/components/claude/ClaudeTokenTracker.vue";
</script>

<template>
  <div class="claude-tokens-page">
    <claude-token-tracker />
  </div>
</template>

<style scoped>
.claude-tokens-page {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}
</style>