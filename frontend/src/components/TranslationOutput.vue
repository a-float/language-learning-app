<script setup lang="ts">
import { type TranslateResponse } from '@/stores/translation'

defineProps<{
  result: TranslateResponse
}>()
</script>

<template>
  <div class="bg-white rounded-lg shadow p-6 space-y-4">
    <h2 class="text-xl font-semibold">Translation</h2>

    <div v-if="result.cached" class="text-sm text-amber-600 bg-amber-50 p-2 rounded">
      ⚡ Cached result
    </div>

    <div class="space-y-2">
      <p class="text-sm text-gray-500">Source</p>
      <p class="text-lg">{{ result.translatedText }}</p>
    </div>
  </div>
</template>
