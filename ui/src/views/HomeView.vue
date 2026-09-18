<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { type LearningMaterialsResponse } from '@/generated/pocketbase-types'
import { pb } from '@/lib/pocketbase'
import { ref, watch } from 'vue'

const { isAuthenticated } = useAuth()
const loading = ref(false)
const data = ref<LearningMaterialsResponse[] | null>(null)
const error = ref<string | null>(null)

watch(isAuthenticated, fetchRecords, { immediate: true })

async function fetchRecords() {
  error.value = data.value = null
  loading.value = true

  try {
    data.value = await pb
      .collection('learningMaterials')
      .getFullList({ page: 1, perPage: 20, expand: 'words,phrases' })
  } catch (err) {
    error.value = String(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <h1>Hello this is the homepage</h1>
  <div v-if="error">{{ error }}</div>
  <div v-if="loading">Loading...</div>
  <div v-for="material in data" style="display: flex; flex-direction: column">
    <span>Name: {{ material.name }}</span>
    <span>Words: {{ material.words.length }}</span>
    <span>Phrases: {{ material.phrases.length }}</span>
    <RouterLink :to="{ name: 'learn', params: { id: material.id } }">Learn</RouterLink>
  </div>
</template>
