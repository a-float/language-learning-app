<script setup lang="ts>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTranslationStore } from '@/stores/translation'
import { useLessonsStore } from '@/stores/lessons'
import { useFavoritesStore } from '@/stores/favorites'
import TranslationInput from '@/components/TranslationInput.vue'
import TranslationOutput from '@/components/TranslationOutput.vue'
import VocabularyList from '@/components/VocabularyList.vue'

const router = useRouter()
const translationStore = useTranslationStore()
const lessonsStore = useLessonsStore()
const favoritesStore = useFavoritesStore()

const savingLesson = ref(false)
const lessonTitle = ref('')
const showSaveDialog = ref(false)

async function saveAsLesson() {
  if (!translationStore.result || !translationStore.lastSourceText) return

  savingLesson.value = true
  try {
    const lesson = await lessonsStore.createLesson({
      title: lessonTitle.value || 'Untitled Lesson',
      sourceText: translationStore.lastSourceText,
      targetLang: translationStore.lastTargetLang || 'es',
      difficulty: 'beginner',
      sourceType: 'text',
    })
    showSaveDialog.value = false
    lessonTitle.value = ''
    router.push({ name: 'song', params: { id: lesson.id } })
  } catch (err) {
    console.error('Failed to save lesson:', err)
  } finally {
    savingLesson.value = false
  }
}

async function addToFavorites(vocabId: string, vocab: { word: string; translation: string }) {
  try {
    await favoritesStore.add(vocabId)
  } catch (err) {
    console.error('Failed to add favorite:', err)
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-indigo-600 text-white p-4">
      <div class="max-w-3xl mx-auto flex items-center justify-between">
        <h1 class="text-xl font-bold">🐾 Mówik</h1>
        <nav class="flex items-center gap-4 text-sm">
          <router-link to="/lessons" class="hover:opacity-80">📚 Lessons</router-link>
          <router-link to="/favorites" class="hover:opacity-80">❤️ Favorites</router-link>
          <router-link to="/quiz" class="hover:opacity-80">🎯 Quiz</router-link>
        </nav>
      </div>
    </header>

    <main class="max-w-3xl mx-auto p-6 space-y-6">
      <TranslationInput @submit="translationStore.translate" />

      <Transition name="fade">
        <div v-if="translationStore.result" class="mt-6 space-y-4">
          <TranslationOutput :result="translationStore.result" />
          <VocabularyList :items="translationStore.vocabulary" @add-favorite="addToFavorites" />

          <!-- Save as Lesson Button -->
          <div class="bg-white rounded-lg shadow p-6">
            <button
              @click="showSaveDialog = true"
              class="w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 transition font-medium"
            >
              💾 Save as Lesson
            </button>
          </div>
        </div>
      </Transition>

      <!-- Save Lesson Dialog -->
      <Transition name="fade">
        <div v-if="showSaveDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div class="bg-white rounded-xl p-6 max-w-md w-full">
            <h3 class="text-xl font-semibold mb-4">Save as Lesson</h3>
            <input
              v-model="lessonTitle"
              type="text"
              placeholder="Lesson title (optional)"
              class="w-full border rounded-md p-3 mb-4 focus:ring-2 focus:ring-indigo-500"
            />
            <div class="flex gap-3">
              <button
                @click="showSaveDialog = false"
                class="flex-1 bg-gray-200 text-gray-700 py-2 rounded-lg hover:bg-gray-300 transition"
              >
                Cancel
              </button>
              <button
                @click="saveAsLesson"
                :disabled="savingLesson"
                class="flex-1 bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 transition disabled:opacity-50"
              >
                {{ savingLesson ? 'Saving...' : 'Save' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </main>
  </div>
</template>