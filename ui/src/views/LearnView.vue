<script setup lang="ts">
import { type LearningMaterialsResponse } from '@/generated/pocketbase-types'
import { ref } from 'vue'

const loading = ref(false)
const data = ref<LearningMaterialsResponse | null>(null)
const error = ref<string | null>(null)
</script>

<template>
  <h1>Hello this is the LEARN view of {{ $route.params.id }}</h1>
  <div v-if="error">{{ error }}</div>
  <div v-if="loading">Loading...</div>
  <!-- TODO fetch and show something -->
</template>
