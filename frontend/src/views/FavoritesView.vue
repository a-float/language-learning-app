<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFavoritesStore } from '@/stores/favorites'

const router = useRouter()
const favoritesStore = useFavoritesStore()

const loading = ref(false)
const error = ref<string | null>(null)

async function loadFavorites() {
  loading.value = true
  error.value = null
  try {
    await favoritesStore.load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load favorites'
  } finally {
    loading.value = false
  }
}

async function removeFavorite(id: string) {
  try {
    await favoritesStore.remove(id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to remove favorite'
  }
}

function goToQuiz() {
  router.push({ name: 'quiz', query: { source: 'favorites' } })
}

function getLangFlag(lang: string): string {
  if (lang === 'es') return '🇪🇸'
  if (lang === 'it') return '🇮🇹'
  return '🌐'
}

onMounted(() => {
  loadFavorites()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-indigo-600 text-white p-4">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <h1 class="text-xl font-bold">🐾 Mówik</h1>
        <router-link to="/" class="text-white hover:opacity-80">
          ← Home
        </router-link>
      </div>
    </header>

    <main class="max-w-4xl mx-auto p-6 space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-2xl font-bold">Favorites</h2>
        <button
          v-if="favoritesStore.items.length > 0"
          @click="goToQuiz"
          class="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition"
        >
          🎯 Quiz Favorites
        </button>
      </div>

      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <p class="mt-2 text-gray-500">Loading favorites...</p>
      </div>

      <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 p-4 rounded-lg">
        {{ error }}
        <button @click="loadFavorites" class="ml-4 text-indigo-600 hover:underline">Retry</button>
      </div>

      <div v-else-if="favoritesStore.items.length === 0" class="bg-white rounded-lg shadow p-12 text-center">
        <p class="text-gray-500 mb-4">No favorites yet.</p>
        <p class="text-sm text-gray-400 mb-6">Save words from your lessons to build your personal vocabulary collection!</p>
        <router-link to="/" class="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 transition">
          Start Translating
        </router-link>
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="fav in favoritesStore.items"
          :key="fav.id"
          class="bg-white rounded-lg shadow p-4 hover:shadow-md transition flex items-center justify-between"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-3">
              <span class="font-medium text-indigo-700 text-lg">{{ fav.word }}</span>
              <span class="flex items-center gap-1">
                <span>{{ getLangFlag(fav.language) }}</span>
                <span class="text-sm text-gray-500">{{ fav.language === 'es' ? 'Spanish' : 'Italian' }}</span>
              </span>
              <span v-if="fav.partOfSpeech" class="px-2 py-0.5 bg-gray-100 rounded text-xs text-gray-600">{{ fav.partOfSpeech }}</span>
              <span v-if="fav.difficulty" class="px-2 py-0.5 bg-gray-100 rounded text-xs">{{ fav.difficulty }}</span>
            </div>
            <p class="text-indigo-600 font-medium mt-1">{{ fav.translation }}</p>
            <p v-if="fav.example" class="text-sm italic text-gray-500 mt-1">"{{ fav.example }}"</p>
          </div>
          <button
            @click="removeFavorite(fav.id)"
            class="text-gray-400 hover:text-red-500 transition text-2xl ml-4"
            title="Remove from favorites"
          >
            ♡
          </button>
        </div>
      </div>
    </main>
  </div>
</template>