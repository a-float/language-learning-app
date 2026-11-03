<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useLessonsStore } from '@/stores/lessons'

const router = useRouter()
const lessonsStore = useLessonsStore()

const loading = ref(false)
const error = ref<string | null>(null)

async function loadLessons() {
  loading.value = true
  error.value = null
  try {
    await lessonsStore.load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load lessons'
  } finally {
    loading.value = false
  }
}

async function deleteLesson(id: string) {
  if (!confirm('Delete this lesson?')) return
  try {
    await lessonsStore.deleteLesson(id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete lesson'
  }
}

function goToLesson(id: string) {
  router.push({ name: 'song', params: { id } })
}

function goToQuiz(lessonId: string) {
  router.push({ name: 'quiz', query: { source: 'lesson', sourceId: lessonId } })
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function getLangFlag(lang: string): string {
  if (lang === 'es') return '🇪🇸'
  if (lang === 'it') return '🇮🇹'
  return '🌐'
}

onMounted(() => {
  loadLessons()
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
        <h2 class="text-2xl font-bold">My Lessons</h2>
      </div>

      <div v-if="loading" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <p class="mt-2 text-gray-500">Loading lessons...</p>
      </div>

      <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 p-4 rounded-lg">
        {{ error }}
        <button @click="loadLessons" class="ml-4 text-indigo-600 hover:underline">Retry</button>
      </div>

      <div v-else-if="lessonsStore.lessons.length === 0" class="bg-white rounded-lg shadow p-12 text-center">
        <p class="text-gray-500 mb-4">No lessons yet.</p>
        <p class="text-sm text-gray-400 mb-6">Translate some text on the home page and save it as a lesson!</p>
        <router-link to="/" class="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 transition">
          Go Translate
        </router-link>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="lesson in lessonsStore.lessons"
          :key="lesson.id"
          class="bg-white rounded-lg shadow p-6 hover:shadow-md transition"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-3 mb-2">
                <h3 class="text-xl font-semibold text-gray-900 truncate">{{ lesson.title }}</h3>
                <span class="flex items-center gap-1">
                  <span>{{ getLangFlag(lesson.language) }}</span>
                  <span class="text-sm text-gray-500">{{ lesson.language === 'es' ? 'Spanish' : 'Italian' }}</span>
                </span>
                <span class="px-2 py-0.5 bg-gray-100 rounded text-xs">{{ lesson.difficulty || 'beginner' }}</span>
                <span class="px-2 py-0.5 bg-gray-100 rounded text-xs capitalize">{{ lesson.sourceType || 'text' }}</span>
              </div>
              <p class="text-gray-600 text-sm mb-2 line-clamp-2">{{ lesson.sourceText }}</p>
              <div class="flex items-center gap-4 text-sm text-gray-500">
                <span>{{ formatDate(lesson.createdAt) }}</span>
                <span>{{ lesson.vocabulary.length }} words</span>
              </div>
            </div>
            <div class="flex items-center gap-2 ml-4">
              <button
                @click="goToLesson(lesson.id)"
                class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition text-sm"
              >
                📖 View
              </button>
              <button
                @click="goToQuiz(lesson.id)"
                class="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition text-sm"
              >
                🎯 Quiz
              </button>
              <button
                @click="deleteLesson(lesson.id)"
                class="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition text-sm"
              >
                🗑 Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>